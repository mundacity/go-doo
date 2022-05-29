package cli

type InstanceTypeNotRecognised struct{}

type UnableToDetermineQueryTypeError struct{}

func (u *UnableToDetermineQueryTypeError) Error() string {
	return "unable to determine desired query type"
}
