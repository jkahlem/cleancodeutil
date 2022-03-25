package lsp

type WorkDonePrgoressReport struct {
	Kind        string `json:"kind"`
	Cancellable bool   `json:"cancellable,omitEmpty"`
	Message     string `json:"message,omitEmpty"`
	Percentage  int    `json:"percentage,omitEmpty"`
}

type WorkDoneProgressBegin struct {
	WorkDonePrgoressReport
	Title string `json:"title"`
}

type WorkDonePrgoressEnd struct {
	Kind    string `json:"kind"`
	Message string `json:"message,omitEmpty"`
}

type ProgressParams struct {
	Token interface{} `json:"token"` // might be int or string
	Value interface{} `json:"value"`
}
