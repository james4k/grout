package grout

var defaultConfig = M{
	"url": "",
	"collections": M{
		"posts": M{
			"dir":       "_posts",
			"generator": "post",
		},
	},
}
