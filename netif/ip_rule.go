package netif

import "github.com/vishvananda/netlink"

func InitIPRule() error {
	rule := netlink.NewRule()
	rule.Table = 52
	rule.Priority = IP_RULE_PRIORITY
	if exists, err := ipRuleExists(rule); err != nil {
		return err
	} else if !exists {
		if err := netlink.RuleAdd(rule); err != nil {
			return err
		}
	}

	return nil
}

func DeinitIPRule() error {
	rule := netlink.NewRule()
	rule.Table = 52
	rule.Priority = IP_RULE_PRIORITY
	return netlink.RuleDel(rule)
}

func ipRuleExists(rule *netlink.Rule) (bool, error) {
	rules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err != nil {
		return false, err
	}

	for _, r := range rules {
		if r.Table == rule.Table && r.Priority == rule.Priority {
			return true, nil
		}
	}

	return false, nil
}
