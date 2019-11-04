package main

import (
	"encoding/json"
	"imooc-product/common"
	"imooc-product/datamodels"
	"imooc-product/encrypt"
	"imooc-product/rabbitmq"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/kataras/golog"
)

//集群地址,尽量使用内网IP
var hostArray = []string{"127.0.0.1", "127.0.0.1"}

// 本地地址，同样使用内网ip
var localhost = "127.0.0.1"

// 数量控制服务器内网ip
var GetOneIp = "127.0.0.1"
var GetOnePort = "8084"

// 端口号
var port = "8083"

// rabbitMQ
var rabbitMQValidate *rabbitmq.RabbitMQ

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
		return GetDataFromOtherMap(host, r)
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
func GetDataFromOtherMap(host string, request *http.Request) (isOK bool) {
	hostUrL := "http://" + host + port + "/check"
	rsp, body, err := GetCurl(hostUrL, request)
	if err != nil {
		golog.Error(err)
		return false
	}
	if rsp.StatusCode == 200 && string(body) == "true" {
		return true
	}
	return false
}

// 模拟请求访问
func GetCurl(hostUrl string, request *http.Request) (resp *http.Response, body []byte, err error) {
	uidCookie, err := request.Cookie("uid")
	if err != nil {
		golog.Error("获取分布式锁时，获取uid失败")
		return nil, nil, err
	}
	signCookie, err := request.Cookie("sign")
	if err != nil {
		golog.Error("获取分布式锁时，获取sign失败", err)
		return nil, nil, err
	}
	client := http.DefaultClient
	req, err := http.NewRequest("GET", hostUrl, nil)
	if err != nil {
		golog.Error("创建http请求出错", err)
		return
	}
	// 将cookie注入请求
	req.AddCookie(uidCookie)
	req.AddCookie(signCookie)
	// 执行请求动作，并获取响应
	rsp, err := client.Do(req)
	if err != nil {
		golog.Error("发送http请求出错", err)
		return nil, nil, err
	}
	defer rsp.Body.Close()
	//
	body, err = ioutil.ReadAll(rsp.Body)
	if err != nil {
		golog.Error("解析http响应出错", err)
		return
	}
	return
}
func Check(rw http.ResponseWriter, r *http.Request) {
	// 正常的业务逻辑
	golog.Debug("执行check")
	// 获取url中的参数
	queryForm, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil && len(queryForm["productID"]) <= 0 && len(queryForm["productID"][0]) <= 0 {
		rw.Write([]byte("false"))
		return
	}
	productIDstr := queryForm["productID"][0]
	golog.Debug("获取到productID", productIDstr)
	// 获取cookie
	uidCookie, err := r.Cookie("uid")
	if err != nil {
		rw.Write([]byte("false"))
		return
	}
	// 分布式权限验证:去访问对应的主机，
	if !accessControl.GetDistributedRight(r) {
		rw.Write([]byte("false"))
	}
	// 获取数量控制权限，防止秒杀超卖
	hostUrl := "http://" + GetOneIp + GetOnePort + "getOne"
	rsp, body, err := GetCurl(hostUrl, r)
	if err != nil {
		rw.Write([]byte("false"))
		return
	}
	if rsp.StatusCode == 200 {
		if string(body) == "true" {
			// 整合下单
			productID, err := strconv.ParseInt(productIDstr, 10, 64)
			if err != nil {
				golog.Error(productID, "失败")
				rw.Write([]byte("false"))
				return
			}
			userID, err := strconv.ParseInt(uidCookie.Value, 10, 64)
			if err != nil {
				golog.Error(userID, "失败")
				rw.Write([]byte("false"))
				return
			}
			message := datamodels.Message{userID, productID}
			msgByte, err := json.Marshal(message)
			if err != nil {
				golog.Error(message, "失败")
				rw.Write([]byte("false"))
				return
			}
			err = rabbitMQValidate.PublishSimple(string(msgByte))
			if err != nil {
				golog.Error(err, "失败")
				rw.Write([]byte("false"))
				return
			}
			rw.Write([]byte("true"))
		}
	}
	rw.Write([]byte("false"))
	return
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
		golog.Error("加密段被篡改，cookie校验失败")
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
	// 获取新ip地址
	lip, err := common.GetIntranceIP()
	if err != nil {
		golog.Warn("获取新本地ip失败，默认使用127.0.0.1！")
	} else {
		localhost = lip
	}
	golog.Debug("本地ip：", localhost)
	//rabbitMQ全局实例创建
	rabbitMQValidate = rabbitmq.NewSimpleRabbitMQ(
		"imoocProduct",
	)
	defer rabbitMQValidate.Destory()

	// 过滤器实例
	filter := common.NewFilter()
	filter.RegisterFilterUri("/check", Auth)
	http.HandleFunc("/check", filter.Handle(Check))
	err = http.ListenAndServe(":8083", nil)
	if err != nil {
		golog.Error("8083服务出错")

	}
}
