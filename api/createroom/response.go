package createroom

type ResponseRoom struct {
	ID string `json:"id"`
}

type ResponseBody struct {
	Room ResponseRoom `json:"room"`
}
