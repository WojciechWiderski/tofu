package mysql

import (
	"context"
	"fmt"
	"reflect"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/WojciechWiderski/tofu/tconfig"
	"github.com/WojciechWiderski/tofu/tdatabase"
	"github.com/WojciechWiderski/tofu/terror"
	"github.com/WojciechWiderski/tofu/tlogger"
	"github.com/WojciechWiderski/tofu/tmodel"
)

type DB struct {
	db     *gorm.DB
	models *tmodel.Models
}

var dsn = "%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local"

func New(conf tconfig.MySql, models *tmodel.Models) *DB {

	db, err := connectMySql(conf)
	if err != nil {
		panic(err)
	}

	tlogger.Success("My sql connected!")

	return &DB{
		db,
		models,
	}
}

func connectMySql(conf tconfig.MySql) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(fmt.Sprintf(dsn, conf.Username, conf.Password, conf.Address, conf.DatabaseName)), &gorm.Config{})
	if err != nil {
		tlogger.Error(fmt.Sprintf("Connect to mysql terror: %s", err))
		return nil, terror.NewInternalf("gorm.Open()", err)
	}
	tlogger.Info("Connect to mysql...")
	return db, nil
}

func (m *DB) Migrate() error {
	for _, model := range m.models.All {
		if err := m.db.AutoMigrate(&model.In); err != nil {
			tlogger.Error(fmt.Sprintf("Migrate terror for: %s, terror: %s", model.Name, err))
			return terror.NewInternalf(fmt.Sprintf("db.AutoMigrate() - model: %s", model.Name), err)
		}
		model.DB = m
		tlogger.Success(fmt.Sprintf("Migrate success for: %s", model.Name))
	}

	return nil
}

func (m *DB) Add(ctx context.Context, in interface{}) error {
	tx := m.db.Begin()

	if result := tx.Create(in); result.Error != nil {
		tx.Rollback()
		return terror.NewInternalf("tx.Create()", fmt.Errorf(result.Error.Error()))
	}
	tx.Commit()
	return nil
}

func (m *DB) GetOne(ctx context.Context, in interface{}, params tdatabase.ParamRequest) (interface{}, error) {
	tx := m.db.Begin()

	if result := tx.First(&in, fmt.Sprintf("%s = ?", params.By), params.Value); result.Error != nil {
		tx.Rollback()
		return nil, terror.NewInternalf("tx.First()", fmt.Errorf(result.Error.Error()))
	}
	tx.Commit()
	return in, nil
}

func (m *DB) GetMany(ctx context.Context, in interface{}, params tdatabase.ParamRequest) ([]interface{}, error) {
	tx := m.db.Begin()
	var result []any
	rows, err := tx.Model(in).Rows()
	if err != nil {
		tx.Rollback()
		return nil, terror.NewInternalf("tx.Model().Rows()", fmt.Errorf(err.Error()))
	}
	defer rows.Close()
	for rows.Next() {
		newIn := reflect.New(reflect.ValueOf(in).Elem().Type()).Interface()
		err := tx.ScanRows(rows, &newIn)
		if err != nil {
			tx.Rollback()
			return nil, terror.NewInternalf("tx.ScanRows()", fmt.Errorf(err.Error()))
		}
		result = append(result, newIn)
	}

	tx.Commit()
	return result, nil
}

func (m *DB) Update(ctx context.Context, update interface{}, in interface{}, id int) error {
	tx := m.db.Begin()
	in, err := m.GetOne(ctx, in, tdatabase.ParamRequest{
		By:    "id",
		Value: id,
	})
	if err != nil {
		tx.Rollback()
		return terror.Wrap("m.GetOne()", fmt.Errorf(err.Error()))
	}
	if result := tx.Model(in).Updates(update); result.Error != nil {
		tx.Rollback()
		return terror.NewInternalf("tx.Model().Updates()", fmt.Errorf(result.Error.Error()))
	}
	tx.Commit()
	return nil
}

func (m *DB) Delete(ctx context.Context, in interface{}, id int) error {
	tx := m.db.Begin()
	if result := tx.Delete(in, id); result.Error != nil {
		tx.Rollback()
		return terror.NewInternalf("tx.Delete()", fmt.Errorf(result.Error.Error()))
	}
	tx.Commit()
	return nil
}
