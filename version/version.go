package version

import "fmt"

var (
	BuildVersion string
	BuildTime    string
	CommitID     string
)

func ShowLogo(buildVersion, buildTime, commitID string) {
	BuildVersion = buildVersion
	BuildTime = buildTime
	CommitID = commitID

	//版本号
	//fmt.Println("   _____                         \n  / ____|                        \n | (___   __ _  __ _  ___   ___  \n  \\___ \\ / _` |/ _` |/ _ \\ / _ \\ \n  ____) | (_| | (_| | (_) | (_) |\n |_____/ \\__,_|\\__, |\\___/ \\___/ \n                __/ |            \n               |___/             ")
	fmt.Println("\n\x1b[32m_________________________________________________________\n      __                               __     __   ______\n    /    )                             /    /    )   /   \n----\\--------__----__----__----__-----/----/----/---/----\n     \\     /   ) /   ) /   ) /   )   /    /    /   /     \n_(____/___(___(_(___/_(___/_(___/_ _/_ __(____/___/______\n                   /                                     \n               (_ /     \x1b[0m\n")
	fmt.Println("Version   ：", buildVersion)
	fmt.Println("BuildTime ：", buildTime)
	fmt.Println("CommitID  ：", commitID)
	fmt.Println("")
}
func GetVersion() string {
	if BuildVersion == "" {
		BuildVersion = "0.0"
	}
	return BuildVersion
}
func GetBuildTime() string {
	return BuildTime
}
func GetCommitID() string {
	return CommitID
}
