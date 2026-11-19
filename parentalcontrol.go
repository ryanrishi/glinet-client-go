package glinet

import "context"

// ParentalControlService calls the parental-control module: device groups and the rulesets that
// they use. A group has a list of client MACs, a default ruleset, a weekly schedule of rulesets,
// and can have a temporary ruleset ("brief"). A ruleset blocks apps of the app list and domains of
// a block list, or loads its block list from a URL.
//
// The reads GetConfig, GetStatus and ListApps were done on a router (firmware 4.7.0 and 4.8.3)
// with no groups and only the preset rulesets. Each write, GetBrief and GetMode are from the web UI
// code of firmware 4.8.3 (view parentalcontrol) and the archived API reference, and were not
// tested on a router.
//
// The web UI shows a warning that parental control does not work while network acceleration
// (network.get_netnat_config) is on.
//
// The router gives each group, ruleset and schedule an ID. The IDs are strings in the reference;
// the preset rulesets have the IDs "drop" and "accept". Whether an ID stays the same after other
// items are added or removed was not tested.
type ParentalControlService service

const parentalControlModule = "parental-control"

// The IDs of the preset rulesets.
const (
	ParentalControlRuleDrop   = "drop"   // blocks all traffic (name "drop_all")
	ParentalControlRuleAccept = "accept" // blocks nothing (name "accept_all")
)

// The modes of parental-control get_mode and set_mode. The web UI uses the modes only on a router
// with Bark: system get_info software_feature.bark, firmware category "2c" and country code "us".
// Mode 0 means that no mode was chosen; the web UI then shows the mode choice.
const (
	ParentalControlModeDefault = 1 // the parental control of this service
	ParentalControlModeBark    = 2 // Bark (BarkService)
)

// ParentalControlConfig is the result of parental-control get_config.
//
// DropAnonymous blocks the clients that are in no group. AutoUpdate updates the app feature
// library automatically. EnableApp is true when the router has the app feature library: the web
// UI shows AutoUpdate and the apps of a ruleset only then. EnableApp is read-only.
type ParentalControlConfig struct {
	Enabled       bool                   `json:"enable"`
	DropAnonymous bool                   `json:"drop_anonymous"`
	AutoUpdate    bool                   `json:"auto_update"`
	EnableApp     bool                   `json:"enable_app"`
	Rules         []ParentalControlRule  `json:"rules"`
	Groups        []ParentalControlGroup `json:"groups"`
}

// ParentalControlSettings are the values of parental-control set_config.
type ParentalControlSettings struct {
	Enabled       bool `json:"enable"`
	DropAnonymous bool `json:"drop_anonymous"`
	AutoUpdate    bool `json:"auto_update"`
}

// ParentalControlRule is a ruleset.
//
// The router makes ID. Preset is true for the rulesets "drop" and "accept", which the web UI does
// not let the user change or remove. Color is the color of the label in the web UI, for example
// "#4bd9cc". With Manual true, the ruleset blocks the apps in Apps (IDs from ListApps) and the
// entries of Blacklist. With Manual false, the router loads the block list from URL, and
// URLSuccess is the number of entries that it loaded, as a string.
//
// Only the preset rulesets were seen on a router. Apps, Blacklist and URL are from the web UI code
// and the reference. That get_config returns URL is an assumption: the web UI reads it from the
// ruleset.
type ParentalControlRule struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Color      string   `json:"color"`
	Preset     bool     `json:"preset"`
	Manual     bool     `json:"manual"`
	Apps       []int    `json:"apps,omitempty"`
	Blacklist  []string `json:"blacklist,omitempty"`
	URL        string   `json:"url,omitempty"`
	URLSuccess string   `json:"url_success,omitempty"`
}

// ParentalControlRuleSettings are the values of a ruleset that AddRule and SetRule send.
//
// Name must be unique (reference). Apps has app IDs from ListApps; the reference also allows the
// app types 1 to 8, which mean all apps of that type. With Manual true, the call sends Apps and
// Blacklist; with Manual false, it sends Apps and URL.
//
// Blacklist has one entry for each line of the block list of the web UI. The web UI removes the
// characters that are not in [A-Za-z0-9.[\]~:;/?&=%\-#] from each line and sends each line, also
// an empty line. The reference also shows entries in the app feature syntax, for example
// "[tcp;;;www.example.com;;]", and allows at most 32 entries.
type ParentalControlRuleSettings struct {
	Name      string
	Color     string
	Apps      []int
	Manual    bool
	Blacklist []string
	URL       string
}

// ParentalControlGroup is a device group.
//
// The router makes ID. MACs are the MACs of the clients in the group; the web UI compares them
// without letter case. DefaultRule is the ID of the ruleset that applies when no schedule applies.
//
// Not seen on a router: the test router had no groups. The fields are from the web UI code and
// the reference.
type ParentalControlGroup struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	DefaultRule string                    `json:"default_rule"`
	MACs        []string                  `json:"macs"`
	Schedules   []ParentalControlSchedule `json:"schedules,omitempty"`
}

// ParentalControlGroupSettings are the values of a group that AddGroup and SetGroup send. The web
// UI allows a Name of at most 64 characters.
type ParentalControlGroupSettings struct {
	Name        string   `json:"name"`
	DefaultRule string   `json:"default_rule"`
	MACs        []string `json:"macs"`
}

// ParentalControlSchedule is an item of the weekly schedule of a group: from Begin to End on
// Week, the group uses the ruleset Rule.
//
// Week is 0 (Sunday) to 6 (Saturday). Begin and End are "hh:mm". The router makes ID; SetSchedules
// sends ID only when it is not "". The web UI sends the schedules that it read with their ID and
// new schedules without an ID.
//
// The web UI accepts a schedule only when it is at least 15 minutes long and does not overlap
// another schedule of the same day. Its time list ends at "24:00", and it also treats "23:59" as
// the end of a day. Not settled: whether the router returns Begin and End with seconds (the web UI
// cuts them to 5 characters), whether it stores "24:00", and whether ID is a string or a number
// (the reference shows both; this type requires a string).
type ParentalControlSchedule struct {
	ID    string `json:"id,omitempty"`
	Week  int    `json:"week"`
	Begin string `json:"begin"`
	End   string `json:"end"`
	Rule  string `json:"rule"`
}

// ParentalControlStatus is the result of parental-control get_status. TimeValid is false until the
// router time is synchronized. Groups has the state of each group.
type ParentalControlStatus struct {
	TimeValid bool                         `json:"time_valid"`
	Groups    []ParentalControlGroupStatus `json:"groups"`
}

// ParentalControlGroupStatus is the state of a group: Rule is the ID of the ruleset that applies
// now, Brief is true while a temporary ruleset applies, and ScheduleID is the ID of the schedule
// that applies, or "" for the default or the temporary ruleset.
//
// Not seen on a router: the fields are from the web UI code and the reference.
type ParentalControlGroupStatus struct {
	ID         string `json:"id"`
	Rule       string `json:"rule"`
	Brief      bool   `json:"brief"`
	ScheduleID string `json:"schedule_id,omitempty"`
}

// ParentalControlBrief is the temporary ruleset of a group, the result of parental-control
// get_brief. While Enabled is true, the group uses the ruleset RuleID until the time of day Time
// ("hh:mm"; the web UI cuts it to 5 characters), or until StopBrief when ManualStop is true.
//
// Not seen on a router: the fields are from the web UI code and the reference. The reference also
// shows group_id in the result.
type ParentalControlBrief struct {
	Enabled    bool   `json:"enable"`
	ManualStop bool   `json:"manual_stop"`
	Time       string `json:"time,omitempty"`
	RuleID     string `json:"rule_id,omitempty"`
	GroupID    string `json:"group_id,omitempty"`
}

// ParentalControlApp is an app of the app feature library. Type is 1 social, 2 game, 3 video,
// 4 shopping, 5 music, 6 recruitment (job), 7 download or 8 portal.
type ParentalControlApp struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type int    `json:"type"`
}

type parentalControlID struct {
	ID string `json:"id"`
}

// GetConfig returns the configuration, with the rulesets and the groups.
func (s *ParentalControlService) GetConfig(ctx context.Context) (*ParentalControlConfig, error) {
	var res ParentalControlConfig
	if err := s.client.Call(ctx, parentalControlModule, "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetConfig sets the global switches with parental-control set_config
// {"enable", "drop_anonymous", "auto_update"}, as the web UI does.
//
// From the web UI code of firmware 4.8.3, not tested on a router. On a router with Bark, the web UI
// also calls SetMode with ParentalControlModeDefault and Bark SetConfig with enable false after it
// enables parental control, when the mode was ParentalControlModeBark. After it enables parental
// control with no group, the web UI opens the dialog to add a group.
func (s *ParentalControlService) SetConfig(ctx context.Context, settings ParentalControlSettings) error {
	return s.client.Call(ctx, parentalControlModule, "set_config", settings, nil)
}

// GetStatus returns the ruleset that applies now to each group.
func (s *ParentalControlService) GetStatus(ctx context.Context) (*ParentalControlStatus, error) {
	var res ParentalControlStatus
	if err := s.client.Call(ctx, parentalControlModule, "get_status", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// ListApps returns the apps of the app feature library. The web UI calls it only when
// GetConfig returns EnableApp true; the test router returned the list also with EnableApp false.
func (s *ParentalControlService) ListApps(ctx context.Context) ([]ParentalControlApp, error) {
	var res struct {
		Apps []ParentalControlApp `json:"apps"`
	}
	if err := s.client.Call(ctx, parentalControlModule, "get_app_list", nil, &res); err != nil {
		return nil, err
	}

	return res.Apps, nil
}

// AddGroup adds a group with parental-control add_group {"name", "default_rule", "macs"} and
// returns the ID that the router gives it. An empty MACs sends []. The web UI does not send
// schedules with add_group; SetSchedules sets them.
//
// From the web UI code of firmware 4.8.3, not tested on a router. Error -1 means that a required
// parameter is missing (reference).
func (s *ParentalControlService) AddGroup(ctx context.Context, group ParentalControlGroupSettings) (string, error) {
	group.MACs = append([]string{}, group.MACs...)
	var res parentalControlID
	if err := s.client.Call(ctx, parentalControlModule, "add_group", group, &res); err != nil {
		return "", err
	}

	return res.ID, nil
}

// SetGroup sets the name, the default ruleset and the MACs of the group id with parental-control
// set_group {"id", "name", "default_rule", "macs"}, as the group page of the web UI does. It does
// not send the schedules, and the web UI expects that the router keeps them. An empty MACs sends
// [].
//
// From the web UI code of firmware 4.8.3, not tested on a router.
func (s *ParentalControlService) SetGroup(ctx context.Context, id string, group ParentalControlGroupSettings) error {
	params := struct {
		ID string `json:"id"`
		ParentalControlGroupSettings
	}{id, group}
	params.MACs = append([]string{}, group.MACs...)
	return s.client.Call(ctx, parentalControlModule, "set_group", params, nil)
}

// SetSchedules replaces the schedules of the group id with parental-control set_group
// {"id", "schedules"}, as the schedule page of the web UI does. An empty schedules sends [], which
// the web UI sends to remove all schedules.
//
// From the web UI code of firmware 4.8.3, not tested on a router. Error -2 means that a schedule
// misses a required parameter (reference).
func (s *ParentalControlService) SetSchedules(ctx context.Context, id string, schedules []ParentalControlSchedule) error {
	params := struct {
		ID        string                    `json:"id"`
		Schedules []ParentalControlSchedule `json:"schedules"`
	}{id, append([]ParentalControlSchedule{}, schedules...)}
	return s.client.Call(ctx, parentalControlModule, "set_group", params, nil)
}

// RemoveGroup removes the group id with parental-control remove_group {"id"}.
//
// From the web UI code of firmware 4.8.3, not tested on a router.
func (s *ParentalControlService) RemoveGroup(ctx context.Context, id string) error {
	return s.client.Call(ctx, parentalControlModule, "remove_group", parentalControlID{id}, nil)
}

// AddRule adds a ruleset with parental-control add_rule and returns the ID that the router gives
// it. It sends {"name", "color", "apps", "manual": true, "blacklist"} or
// {"name", "color", "apps", "manual": false, "url"}, as the web UI does. An empty Apps or
// Blacklist sends []. The web UI waits up to 30 s for the result.
//
// From the web UI code of firmware 4.8.3, not tested on a router. Error -2 means that the block
// list has a domain that is not valid; the router then also returns a valid_list, which the web UI
// offers as a correction and which this method does not return.
func (s *ParentalControlService) AddRule(ctx context.Context, rule ParentalControlRuleSettings) (string, error) {
	var res parentalControlID
	if err := s.client.Call(ctx, parentalControlModule, "add_rule", ruleParams("", rule), &res); err != nil {
		return "", err
	}

	return res.ID, nil
}

// SetRule replaces the ruleset id with parental-control set_rule. It sends the values of AddRule
// and "id".
//
// From the web UI code of firmware 4.8.3, not tested on a router. The web UI also sends back the
// other fields that it read, such as preset and url_success; this method does not. Errors are as
// in AddRule.
func (s *ParentalControlService) SetRule(ctx context.Context, id string, rule ParentalControlRuleSettings) error {
	return s.client.Call(ctx, parentalControlModule, "set_rule", ruleParams(id, rule), nil)
}

// RemoveRule removes the ruleset id with parental-control remove_rule {"id"}.
//
// From the web UI code of firmware 4.8.3, not tested on a router. Error -2 means that a group uses
// the ruleset as its default ruleset, and -3 that a group uses it as its temporary ruleset
// (reference; the web UI labels -3 "in use"). Whether a schedule that uses the ruleset blocks the
// removal is not known.
func (s *ParentalControlService) RemoveRule(ctx context.Context, id string) error {
	return s.client.Call(ctx, parentalControlModule, "remove_rule", parentalControlID{id}, nil)
}

func ruleParams(id string, rule ParentalControlRuleSettings) any {
	type common struct {
		ID     string `json:"id,omitempty"`
		Name   string `json:"name"`
		Color  string `json:"color"`
		Apps   []int  `json:"apps"`
		Manual bool   `json:"manual"`
	}
	c := common{id, rule.Name, rule.Color, append([]int{}, rule.Apps...), rule.Manual}
	if rule.Manual {
		return struct {
			common
			Blacklist []string `json:"blacklist"`
		}{c, append([]string{}, rule.Blacklist...)}
	}
	return struct {
		common
		URL string `json:"url"`
	}{c, rule.URL}
}

// GetBrief returns the temporary ruleset of the group groupID, with parental-control get_brief
// {"group_id"}.
//
// From the web UI code of firmware 4.8.3, not tested on a router.
func (s *ParentalControlService) GetBrief(ctx context.Context, groupID string) (*ParentalControlBrief, error) {
	params := struct {
		GroupID string `json:"group_id"`
	}{groupID}
	var res ParentalControlBrief
	if err := s.client.Call(ctx, parentalControlModule, "get_brief", params, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// StartBrief makes the group groupID use the ruleset ruleID now, until the time of day stopTime
// ("hh:mm"), with parental-control set_brief
// {"enable": true, "manual_stop": false, "rule_id", "group_id", "time"}. With stopTime "", it sends
// {"enable": true, "manual_stop": true, "rule_id", "group_id"}, and the temporary ruleset applies
// until StopBrief. The web UI proposes a stop time 15 minutes after the router time, and a button
// adds 15 minutes to the stop time.
//
// From the web UI code of firmware 4.8.3, not tested on a router. The web UI reads the status 2 s
// after the call.
func (s *ParentalControlService) StartBrief(ctx context.Context, groupID, ruleID, stopTime string) error {
	params := struct {
		Enabled    bool   `json:"enable"`
		ManualStop bool   `json:"manual_stop"`
		RuleID     string `json:"rule_id"`
		GroupID    string `json:"group_id"`
		Time       string `json:"time,omitempty"`
	}{true, stopTime == "", ruleID, groupID, stopTime}
	return s.client.Call(ctx, parentalControlModule, "set_brief", params, nil)
}

// StopBrief stops the temporary ruleset of the group groupID with parental-control set_brief
// {"enable": false, "group_id"}.
//
// From the web UI code of firmware 4.8.3, not tested on a router.
func (s *ParentalControlService) StopBrief(ctx context.Context, groupID string) error {
	params := struct {
		Enabled bool   `json:"enable"`
		GroupID string `json:"group_id"`
	}{false, groupID}
	return s.client.Call(ctx, parentalControlModule, "set_brief", params, nil)
}

// GetMode returns the parental control mode, ParentalControlModeDefault or ParentalControlModeBark,
// or 0 when no mode was chosen.
//
// From the web UI code of firmware 4.8.3, not tested on a router. The web UI calls it only on a
// router with Bark (see the mode constants); what it returns on other routers is not known.
func (s *ParentalControlService) GetMode(ctx context.Context) (int, error) {
	var res struct {
		Mode int `json:"mode"`
	}
	if err := s.client.Call(ctx, parentalControlModule, "get_mode", nil, &res); err != nil {
		return 0, err
	}

	return res.Mode, nil
}

// SetMode sets the parental control mode with parental-control set_mode {"mode"}.
//
// From the web UI code of firmware 4.8.3, not tested on a router. After it sets
// ParentalControlModeDefault, the web UI calls Bark SetConfig with enable false. The web UI calls
// SetMode only on a router with Bark.
func (s *ParentalControlService) SetMode(ctx context.Context, mode int) error {
	params := struct {
		Mode int `json:"mode"`
	}{mode}
	return s.client.Call(ctx, parentalControlModule, "set_mode", params, nil)
}
