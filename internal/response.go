package internal

type Response struct {
	Code int
	Msg  string
}

type UserInfo struct {
	Response
	Name string
	Age  int
}
