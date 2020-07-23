package rbac

func GetUserClusterList(username string) []string {
	userClusterMap := map[string][]string{
		"user1": []string{"cluster1"},
		"user2": []string{"cluster2"},
		"admin": []string{"hub_cluster", "cluster1", "cluster2"},
	}

	return userClusterMap[username]
}
