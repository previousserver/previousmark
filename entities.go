package main

type msg struct {
	Body string `json:"error"`
}

type benchmarks struct {
	Benchmarks []benchmark `json:"benchmarks"`
	Previous string `json:"previous"`
	Next string `json:"next"`
	NewToken token `json:"token"`
}

type benchmark struct {
	BID string `json:"bID"`
	Title string `json:"title"`
	Icon string `json:"iconUrl"`
	Description string `json:"description"`
	Metric string `json:"metric"`
	Rating float32 `json:"rating"`
	RatingCount int `json:"ratingCount"`
	NewToken token `json:"token"`
}

type benchmarkComments struct {
	BenchmarkComments []benchmarkComment `json:"benchmarkComments"`
	Benchmark benchmark `json:"benchmark"`
	User user `json:"user"`
	Previous string `json:"previous"`
	Next string `json:"next"`
	NewToken token `json:"token"`
}

type benchmarkComment struct {
	CID string `json:"cID"`
	Body string `json:"body"`
	Benchmark benchmark `json:"benchmark"`
	User user `json:"user"`
	NewToken token `json:"token"`
}

type submissions struct {
	Submissions []submission `json:"submissions"`
	Benchmark benchmark `json:"benchmark"`
	User user `json:"user"`
	Previous string `json:"previous"`
	Next string `json:"next"`
	NewToken token `json:"token"`
}

type submission struct {
	SID string `json:"sID"`
	Result float32 `json:"result"`
	Screenshot string `json:"screenshotUrl"`
	Rating float32 `json:"rating"`
	RatingCount int `json:"ratingCount"`
	IsVerified bool `json:"verified"`
	Benchmark benchmark `json:"benchmark"`
	User user `json:"user"`
	NewToken token `json:"token"`
}

type submissionComments struct {
	SubmissionComments []submissionComment `json:"submissionComments"`
	Submission submission `json:"submission"`
	User user `json:"user"`
	Previous string `json:"previous"`
	Next string `json:"next"`
	NewToken token `json:"token"`
}

type submissionComment struct {
	CID string `json:"cID"`
	Body string `json:"body"`
	Submission submission `json:"submission"`
	User user `json:"user"`
	NewToken token `json:"token"`
}

type users struct {
	Users []user `json:"users"`
	Previous string `json:"previous"`
	Next string `json:"next"`
	NewToken token `json:"token"`
}

type user struct {
	ID string `json:"ID"`
	Email string `json:"email"`
	Nickname string `json:"nickname"`
	Password string `json:"password"`
	Avatar string `json:"avatarUrl"`
	IsBanned bool `json:"banned"`
	IsVerified bool `json:"verified"`
	IsMod bool `json:"mod"`
	NewToken token `json:"token"`
}

type token struct {
	Token string `json:"token"`
	Status msg `json:"status"`
}