package main

import "log"

func main() {
	c := client{}
	c.login("", "")
	log.Printf("%+v\n", c)

	cid := c.getCustomerID()
	log.Printf("%s\n", cid)

	dd := c.getDistributionGroups(cid)
	log.Printf("%+v\n", dd)

	for _, d := range dd {
		m := c.gatherMetrics(d.ID)
		log.Printf("%+v\n", m)
	}
}
