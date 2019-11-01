package main

import (
	"imooc-product/common"
	"imooc-product/encrypt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/kataras/golog"
)

//集群地址,尽量使用内网IP
var hostArray = []string{"127.0.0.1", "127.0.0.1"}

// 本地地址，同样使用内网ip
var localhost = "127.0.0.1"

// 端口号
var port = "8081"

// 一个全局一致性实例
var hashConsistent *common.Consistent

// 定义一个结构体用来存放控制信息
type AccessControl struct {
	// 结构体包含一个map字段和读写锁
	sourceArray map[int]interface{}
	sync.RWMutex
}

// 控制信息全局实例
var accessControl = &AccessControl{sourceArray: make(map[int]interface{})}

// 获取指定的数据
func (a *AccessControl) GetNewRecord(uid int) interface{} {
	a.RLock()
	defer a.RUnlock()
	return a.sourceArray[uid]
}

// 添加一条数据
func (a *AccessControl) SetNewRecord(uid int) {
	a.Lock()
	defer a.Unlock()
	a.sourceArray[uid] = "hello imooc"
}

// 获取分布式锁
func (a *AccessControl) GetDistributedRight(r *http.Request) bool {
	// 获取发来请求的用户身份id
	uidCookie, err := r.Cookie("uid")
	if err != nil {
		golog.Error("获取分布式锁时，获取uid失败")
		return false
	}
	uid, err := strconv.Atoi(uidCookie.Value)
	if err != nil {
		golog.Error("uid格式错误", err)
		return false
	}
	signCookie, err := r.Cookie("sign")
	if err != nil {
		golog.Error("获取分布式锁时，获取sign失败", err)
		return false
	}
	// 采用一致性hash算法，确定改用户应该访问的机器
	host, err := hashConsistent.Get(uidCookie.Value)
	if err != nil {
		golog.Error("获取节点时出错", err)
		return false
	}
	// 如果目标机器是本机
	if host == localhost {
		return a.GetDataFromMap(uid)
	} else {
		// 非本机，则本机代理请求目标机器是否成功
		return GetDataFromOtherMap(host, uidCookie, signCookie)
	}
}

// 获取本机map
func (a *AccessControl) GetDataFromMap(uid int) (isOK bool) {
	data := a.GetNewRecord(uid)
	if data == nil {
		isOK = false
	}
	return true
}

// 获取其他节点的map处理结果
func GetDataFromOtherMap(host string, uidCookie, signCookie *http.Cookie) (isOK bool) {
	// 模拟请求访问
	client := http.DefaultClient
	req, err := http.NewRequest("GET", "http://"+host+port+"/check", nil)
	if err != nil {
		golog.Error("创建http请求出错", err)
		return false
	}
	// 将cookie注入请求
	req.AddCookie(uidCookie)
	req.AddCookie(signCookie)
	// 执行请求动作，并获取响应
	rsp, err := client.Do(req)
	if err != nil {
		golog.Error("发送http请求出错", err)
		return false
	}
	//
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		golog.Error("解析http响应出错", err)
		return false
	}
	if rsp.StatusCode == 200 && string(body) == "true" {
		return true
	}
	return false
}
func Check(rw http.ResponseWriter, r *http.Request) {
	// 正常的业务逻辑
	golog.Debug("执行check")
	accessControl.GetDistributedRight(r)
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
	// 负载均衡器：给分布式hash算法添加节点
	hashConsistent = common.NewConsistent()
	for _, v := range hostArray {
		hashConsistent.Add(v)
	}
	// 过滤器实例
	filter := common.NewFilter()
	filter.RegisterFilterUri("/check", Auth)
	http.HandleFunc("/check", filter.Handle(Check))
	err := http.ListenAndServe(":8083", nil)
	if err != nil {
		golog.Debug("8083服务出错")

	}
}
