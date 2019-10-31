package repositories

import (
	"database/sql"
	"errors"
	"imooc-product/common"
	"imooc-product/datamodels"

	"github.com/kataras/golog"
)

// 用户表逻辑，面向数据库
type IUserRepository interface {
	Conn() error
	Select(userName string) (*datamodels.User, error)
	Insert(*datamodels.User) (int64, error)
	SelectByID(ID int64) (*datamodels.User, error)
}

type UserManagerRepository struct {
	table     string
	mysqlConn *sql.DB
}

func NewUserManagerRepository(table string, mysqlConn *sql.DB) IUserRepository {
	return &UserManagerRepository{table, mysqlConn}
}

func (u *UserManagerRepository) Conn() error {
	if u.table == "" {
		u.table = "user"
	}
	if u.mysqlConn == nil {
		db, err := common.NewMysqlConn()
		if err != nil {
			return err
		}
		u.mysqlConn = db
	}
	return nil
}
func (u *UserManagerRepository) Select(userName string) (*datamodels.User, error) {
	if u.Conn() != nil {
		return &datamodels.User{}, errors.New("连接已经关闭！")
	}
	if u.table == "" {
		u.table = "user"
	}
	if userName == "" {
		return &datamodels.User{}, errors.New("用户名不能为空！")
	}
	sql := "SELECT * FROM " + u.table + " WHERE userName = ?"
	rows, err := u.mysqlConn.Query(sql, userName)
	if err != nil {
		golog.Error(err)
		return &datamodels.User{}, err
	}
	defer rows.Close()
	// 将查询结果解析出来
	mm := common.GetResultRows(rows)
	if len(mm) == 0 {
		return &datamodels.User{}, errors.New("用户不存在！")
	}
	user := new(datamodels.User)
	common.DataToStructByTagSql(mm[0], user)
	return user, nil
}

// 插入用户记录，成功返回ID,失败返回-1和错误
func (u *UserManagerRepository) Insert(user *datamodels.User) (int64, error) {
	if u.Conn() != nil {
		return -1, errors.New("连接已经关闭！")
	}
	if u.table == "" {
		u.table = "user"
	}
	if user == nil {
		return -1, errors.New("user结构体不能为nil")
	}
	sql := "INSERT " + u.table + " SET nickName=?, userName=?, password=? "
	stmt, err := u.mysqlConn.Prepare(sql)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(user.NickName, user.UserName, user.HashPassword)
	if err != nil {
		return -1, err
	}
	return res.RowsAffected()
}

// 通过ID查询用户记录
func (u *UserManagerRepository) SelectByID(ID int64) (*datamodels.User, error) {
	if u.Conn() != nil {
		return &datamodels.User{}, errors.New("连接已经关闭！")
	}
	if u.table == "" {
		u.table = "user"
	}
	// sql := "SELECT * FROM " + u.table + " WHERE ID = " + strconv.FormatInt(ID, 10)
	// rows, err := u.mysqlConn.Query(sql)
	sql := "SELECT * FROM " + u.table + " WHERE ID = "
	rows, err := u.mysqlConn.Query(sql, ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := common.GetResultRow(rows)
	if len(m) == 0 {
		return nil, errors.New("未找到用户记录！")
	}
	user := new(datamodels.User)
	common.DataToStructByTagSql(m, user)
	return user, nil
}
