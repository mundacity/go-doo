package cli

type InstanceTypeNotRecognised struct{}

func (i *InstanceTypeNotRecognised) Error() string {
	return "invalid instance type; valid options = local(0), remote(1), multiple(2)"
}

type UnableToDetermineQueryTypeError struct{}

func (u *UnableToDetermineQueryTypeError) Error() string {
	return "unable to determine desired query type"
}
