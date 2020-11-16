package fetcher

import (
	"sort"
)

// Create an object with all the domain, and inside an array,
// all the sites object that belong to him.
type SitesPerDomain struct {
	Domain string
	Name   string
	Sites  []SiteCertProber
}

// Receive a set of sites certificates, create [ ]SitesPerDomain to store the sites,
// extract the domain with the publicsuffix.Name.
// Index sites by domain and inside each domain sort them by remainrecipientConfiging certificate days in ascending order
// and return it.
func IndexSitesPerDomains(sitesList []SiteCertProber) []SitesPerDomain {
	domainsSorted := make([]SitesPerDomain, 0)
	domainExists := false
	sitesList[0].GetConfig()
	for _, site := range sitesList {
		for index, domainSorted := range domainsSorted {
			if site.GetDomain() == domainSorted.Name {
				domainExists = true
				domainsSorted[index].Sites = append(domainsSorted[index].Sites, site)
			}
		}
		if domainExists == false {
			domainSorted := SitesPerDomain{
				Name:  site.GetDomain(),
				Sites: []SiteCertProber{site},
			}
			domainsSorted = append(domainsSorted, domainSorted)
		}
	}
	for _, domainSorted := range domainsSorted {
		sort.SliceStable(domainSorted.Sites, func(i, j int) bool {
			return domainSorted.Sites[i].DaysLeft() < domainSorted.Sites[j].DaysLeft()
		})
	}
	return domainsSorted
}
