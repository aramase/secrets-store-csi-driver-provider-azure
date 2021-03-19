package KuberenetesAPIServer

type AlreadyExistError struct{
}

func(error *AlreadyExistError)Error() string{
	return "Resource Already Exist"
}