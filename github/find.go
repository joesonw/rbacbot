package github

import (
	"fmt"

	"github.com/gobwas/glob"
)

func findRelatedUsers(config *Config, changes []string) (map[string]bool, error) {
	users := make(map[string]bool)
	// iterate through each set of access rules
	for file, roles := range config.Access {
		g, err := glob.Compile(file)
		if err != nil {
			return nil, fmt.Errorf("unable to parse glob: %s", err.Error())
		}
		// for each changed file
		for _, change := range changes {
			if g.Match(change) {
				// find users who have roles match either of current access rule required roles(one-of)
				for _, accessRole := range roles {
					for role, members := range config.Roles {
						if role == accessRole {
							for _, member := range members {
								if member.User != "" {
									users[member.User] = true
								}
							}
						}
					}
				}
				break
			}
		}
	}
	return users, nil
}
