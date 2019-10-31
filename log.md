# 项目日志

## 拼SQL语句

一定要注意空格，否则可能拼出来无法执行

## 解析表单数据到结构体

```go
	// 创建空user实例
	user := new(datamodels.User)
	// 解析前，需要转换请求里的表单，先给Form字段注入数据。
	u.Ctx.Request().ParseForm()
	// 创建解码器，设置tag
	decoder := common.NewDecoder(&common.DecoderOptions{TagName: "form"})
	// 使用解码器，将表单数据，解析到空实例中
	err := decoder.Decode(u.Ctx.Request().Form, user)
	if err != nil {
		golog.Error(err)
		u.Ctx.Redirect("/user/error")
		return
	}
	golog.Debug("根据用户表单，得到user实例：", user)
```