package tofu

import (
	"context"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySqlDB struct {
	db     *gorm.DB
	models *Models
}

var dsn = "%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local"

func NewMySqlDB(conf MySqlConfig, models *Models) *MySqlDB {

	db, err := connectMySql(conf)
	if err != nil {
		panic(err)
	}

	return &MySqlDB{
		db,
		models,
	}
}

func connectMySql(conf MySqlConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(fmt.Sprintf(dsn, conf.Username, conf.Password, conf.Address, conf.DatabaseName)), &gorm.Config{})
	if err != nil {
		return nil, NewInternalf("gorm.Open()", err)
	}
	return db, nil
}

func (m *MySqlDB) Migrate() error {
	for _, model := range m.models.All {
		if err := m.db.AutoMigrate(&model.In); err != nil {
			return NewInternalf(fmt.Sprintf("db.AutoMigrate() - model: %s", model.Name), err)
		}
	}
	return nil
}

func (m *MySqlDB) Add(ctx context.Context, in interface{}) error {
	tx := m.db.Begin()
	if result := m.db.Create(in); result.Error != nil {
		tx.Rollback()
		return NewInternalf("ms.Create(in)", fmt.Errorf(result.Error.Error()))
	}
	tx.Commit()
	return nil
}

func (m *MySqlDB) Get(ctx context.Context, in interface{}, params ParamRequest) (interface{}, error) {
	tx := m.db.Begin()

	if result := m.db.First(&in, fmt.Sprintf("%s = ?", params.By), params.Value); result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}
	tx.Commit()
	return in, nil
}

func (m *MySqlDB) Update(ctx context.Context, update interface{}, in interface{}, id int) error {
	tx := m.db.Begin()
	in, err := m.Get(ctx, in, ParamRequest{
		By:    "id",
		Value: id,
	})
	if err != nil {
		tx.Rollback()
		return err
	}
	if result := m.db.Model(in).Updates(update); result.Error != nil {
		tx.Rollback()
		return result.Error
	}
	tx.Commit()
	return nil
}
