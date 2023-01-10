package main

type msg struct {
	Body string `json:"error"`
}

type benchmarks struct {
	Benchmarks []benchmark `json:"benchmarks"`
	Previous   string      `json:"previous"`
	Next       string      `json:"next"`
	NewToken   token       `json:"token"`
}

type benchmark struct {
	BID         string `json:"bid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Url         string `json:"url"`
	NewToken    token  `json:"token"`
}

type blogposts struct {
	Blogposts []blogpost `json:"blogposts"`
	Previous  string     `json:"previous"`
	Next      string     `json:"next"`
	NewToken  token      `json:"token"`
}

type blogpost struct {
	BPID       string   `json:"bpid"`
	Title      string   `json:"title"`
	Body       string   `json:"body"`
	User       user     `json:"user"`
	Created    string   `json:"created"`
	IsVerified bool     `json:"verified"`
	Tags       []string `json:"tags"`
	NewToken   token    `json:"token"`
}

type blogpostComments struct {
	BlogpostComments []blogpostComment `json:"blogpostComments"`
	Blogpost         blogpost          `json:"blogpost"`
	User             user              `json:"user"`
	Previous         string            `json:"previous"`
	Next             string            `json:"next"`
	NewToken         token             `json:"token"`
}

type blogpostComment struct {
	BCID       string   `json:"bcid"`
	Body       string   `json:"body"`
	Blogpost   blogpost `json:"blogpost"`
	User       user     `json:"user"`
	Created    string   `json:"created"`
	IsVerified bool     `json:"verified"`
	NewToken   token    `json:"token"`
}

type submissions struct {
	Submissions []submission `json:"submissions"`
	Benchmark   benchmark    `json:"benchmark"`
	User        user         `json:"user"`
	Previous    string       `json:"previous"`
	Next        string       `json:"next"`
	NewToken    token        `json:"token"`
}

type submission struct {
	SID        string    `json:"sid"`
	Result     float32   `json:"result"`
	Screenshot string    `json:"screenshot"`
	Processor  processor `json:"processor"`
	Memory     memory    `json:"memory"`
	MemCount   string    `json:"memCount"`
	Created    string    `json:"created"`
	IsVerified bool      `json:"verified"`
	Url        string    `json:"url"`
	Benchmark  benchmark `json:"benchmark"`
	User       user      `json:"user"`
	Post       blogpost  `json:"post"`
	NewToken   token     `json:"token"`
}

type processor struct {
	PID          string `json:"pid"`
	Model        string `json:"model"`
	Lineup       string `json:"lineup"`
	Manufacturer string `json:"manufacturer"`
	NewToken     token  `json:"token"`
}

type memory struct {
	MID          string `json:"mid"`
	Model        string `json:"model"`
	Generation   string `json:"generation"`
	Capacity     string `json:"capacity"`
	Lineup       string `json:"lineup"`
	Manufacturer string `json:"manufacturer"`
	NewToken     token  `json:"token"`
}

type submissionComments struct {
	SubmissionComments []submissionComment `json:"submissionComments"`
	Submission         submission          `json:"submission"`
	User               user                `json:"user"`
	Previous           string              `json:"previous"`
	Next               string              `json:"next"`
	NewToken           token               `json:"token"`
}

type submissionComment struct {
	SCID       string     `json:"scid"`
	Body       string     `json:"body"`
	Submission submission `json:"submission"`
	User       user       `json:"user"`
	NewToken   token      `json:"token"`
}

type users struct {
	Users    []user `json:"users"`
	Previous string `json:"previous"`
	Next     string `json:"next"`
	NewToken token  `json:"token"`
}

type user struct {
	UID        string `json:"uid"`
	Email      string `json:"email"`
	Nickname   string `json:"nickname"`
	Nuke       string `json:"nuke"`
	Avatar     string `json:"avatar"`
	AboutMe    string `json:"aboutMe"`
	AboutBlog  string `json:"aboutBlog"`
	IsVerified bool   `json:"verified"`
	IsMod      bool   `json:"mod"`
	Created    string `json:"created"`
	LastNuke   string `json:"lastNuke"`
	NewToken   token  `json:"token"`
}

type token struct {
	Token  string `json:"token"`
	Status msg    `json:"status"`
}
