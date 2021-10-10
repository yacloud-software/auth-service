package defgroups

import (
	pb "golang.conradwood.net/apis/auth"
)

// add default groups, but only those that aren't added yet
func AddDefaultGroups(a *pb.User) {
	defGroups := GetDefaultGroups()
	var add []*pb.Group
	for _, dg := range defGroups {
		found := false
		if a.Groups != nil {

			for _, ag := range a.Groups {
				if ag.ID == dg.ID {
					found = true
					break
				}
			}

		}
		if !found {
			add = append(add, dg)
		}
	}
	a.Groups = append(a.Groups, add...)
}
func GetDefaultGroups() []*pb.Group {
	var res []*pb.Group
	res = append(res, &pb.Group{ID: "all", Name: "AllUsers"})
	return res
}
func stripDefaultGroups(a *pb.User) {
	var r []*pb.Group
	for _, g := range a.Groups {
		if g.ID == "all" {
			continue
		}
		r = append(r, g)
	}
	a.Groups = r
}

func ConvertUserToResponse(u *pb.User) *pb.User {
	AddDefaultGroups(u)
	return u
}
