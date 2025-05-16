package gemini

type Request struct {
	SystemInstruction Instruction `json:"system_instruction"`
	Contents          Instruction `json:"contents"`
}

type Instruction struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type Response struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Instruction `json:"content"`
}
