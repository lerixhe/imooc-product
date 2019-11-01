package main

import (
	"imooc-product/common"
	"imooc-product/encrypt"
	"net/http"

	"github.com/kataras/golog"
)

func Check(rw http.ResponseWriter, r *http.Request) {
	// 正常的业务逻辑
	golog.Debug("执行check")
}

// 统一验证拦截器，每个接口都需要提前验证
func Auth(rw http.ResponseWriter, r *http.Request) error {
	golog.Debug("执行验证")
	err := CheckUserInfo(r)
	if err != nil {
		golog.Error("验证失败")
		return err
	}
	return nil
}

// 校验用户信息函数
func CheckUserInfo(r *http.Request) error {
	userID, err := r.Cookie("uid")
	if err != nil {
		golog.Error("用户uid Cookie获取失败！")
		return err
	}
	signKey, err := r.Cookie("sign")
	if err != nil {
		golog.Error("用户sign Cookie获取失败！")
		return err
	}
	if userID == nil || signKey == nil {
		golog.Error("用户 Cookie不能为nil ！")
		return err
	}
	strByte, err := encrypt.DePwdCode(signKey.Value)
	if err != nil {
		golog.Error("cookie解密失败")
		return err
	}
	golog.Debug("用户ID", userID.Value)
	golog.Debug("解密后ID", string(strByte))
	if userID.Value != string(strByte) {
		golog.Error("用户校验失败")
		return err
	}
	golog.Debug("当前用户校验成功", userID.Value)
	return nil
}
func main() {
	golog.SetLevel("debug")
	// 过滤器实例
	filter := common.NewFilter()
	filter.RegisterFilterUri("/check", Auth)
	http.HandleFunc("/check", filter.Handle(Check))
	err := http.ListenAndServe(":8083", nil)
	if err != nil {
		golog.Debug("8083服务出错")

	}
}
