package authbe

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

var (
	sudoers_flag = flag.String("sudoers", "", "comma delimited string of userids allowed to sudo")
)

func GetSudoers() []string {
	if *sudoers_flag == "" {
		return nil
	}
	var res []string
	for _, sd := range strings.Split(*sudoers_flag, ",") {
		sd = strings.Trim(sd, " ")
		_, err := strconv.Atoi(sd)
		if err != nil {
			fmt.Printf("Skipping invalid sudoers \"%s\"\n", sd)
			continue
		}
		res = append(res, sd)
	}
	return res

}
