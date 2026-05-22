package ui

import (
	"strings"

	tui "github.com/grindlemire/go-tui"
)

type homeScreenModel struct {
	items              []HomeServerItem
	groupNames         []string
	groupIndexByName   map[string]int
	serverCountByGroup map[string]int
	firstServerByGroup map[string]int
	selectedGroup      int
	selectedServer     int
}

func newHomeScreenModel(items []HomeServerItem, groups []string) *homeScreenModel {
	normalizedItems := normalizeHomeItems(items)
	groupNames := buildHomeGroupNames(normalizedItems, groups)

	model := &homeScreenModel{
		items:              normalizedItems,
		groupNames:         groupNames,
		groupIndexByName:   buildGroupIndexByName(groupNames),
		serverCountByGroup: buildServerCountByGroup(normalizedItems),
		firstServerByGroup: buildFirstServerByGroup(normalizedItems),
		selectedServer:     -1,
	}
	if len(normalizedItems) > 0 {
		model.selectedServer = 0
		model.syncGroupFromServer()
	}

	debugf("home groups initialized names=%v selectedGroup=%d selectedServer=%d", model.groupNames, model.selectedGroup, model.selectedServer)
	return model
}

func normalizeHomeItems(items []HomeServerItem) []HomeServerItem {
	normalized := make([]HomeServerItem, 0, len(items))
	for _, item := range items {
		item.Group = normalizeGroupName(item.Group)
		normalized = append(normalized, item)
	}
	return normalized
}

func buildHomeGroupNames(items []HomeServerItem, groups []string) []string {
	groupNames := make([]string, 0, len(groups)+len(items))
	seenGroups := map[string]struct{}{}

	for _, item := range items {
		if _, exists := seenGroups[item.Group]; exists {
			continue
		}
		seenGroups[item.Group] = struct{}{}
		groupNames = append(groupNames, item.Group)
	}

	for _, group := range groups {
		normalizedGroup := normalizeGroupName(group)
		if _, exists := seenGroups[normalizedGroup]; exists {
			continue
		}
		seenGroups[normalizedGroup] = struct{}{}
		groupNames = append(groupNames, normalizedGroup)
	}

	if len(groupNames) == 0 {
		return []string{"Default"}
	}
	return groupNames
}

func buildGroupIndexByName(groups []string) map[string]int {
	indexByName := make(map[string]int, len(groups))
	for idx, group := range groups {
		indexByName[group] = idx
	}
	return indexByName
}

func buildServerCountByGroup(items []HomeServerItem) map[string]int {
	counts := map[string]int{}
	for _, item := range items {
		counts[item.Group]++
	}
	return counts
}

func buildFirstServerByGroup(items []HomeServerItem) map[string]int {
	firstServerByGroup := map[string]int{}
	for idx, item := range items {
		if _, exists := firstServerByGroup[item.Group]; exists {
			continue
		}
		firstServerByGroup[item.Group] = idx
	}
	return firstServerByGroup
}

func normalizeGroupName(group string) string {
	trimmed := strings.TrimSpace(group)
	if trimmed == "" {
		return "Default"
	}
	return trimmed
}

func (m *homeScreenModel) entryIndex() int {
	if len(m.items) == 0 || m.selectedServer < 0 || m.selectedServer >= len(m.items) {
		return -1
	}
	return m.items[m.selectedServer].EntryIndex
}

func (m *homeScreenModel) moveUp() {
	if m.selectedServer <= 0 {
		return
	}
	m.selectedServer--
	m.syncGroupFromServer()
}

func (m *homeScreenModel) moveDown() {
	if len(m.items) == 0 {
		return
	}
	if m.selectedServer < 0 {
		m.selectedServer = 0
		m.syncGroupFromServer()
		return
	}
	if m.selectedServer >= len(m.items)-1 {
		return
	}
	m.selectedServer++
	m.syncGroupFromServer()
}

func (m *homeScreenModel) advanceGroup(step int) {
	nextGroup := nextGroupIndexWithServers(m.groupNames, m.serverCountByGroup, m.selectedGroup, step)
	if nextGroup < 0 {
		return
	}
	m.selectedGroup = nextGroup
	m.syncServerFromGroup()
}

func (m *homeScreenModel) render(overviewList *tui.Element) {
	m.clampSelectedServer()
	if len(m.items) > 0 {
		m.syncGroupFromServer()
	}

	rows := buildHomeOverviewRows(m.items, m.groupNames)
	debugf("home render overview rows=%d selectedGroup=%d selectedServer=%d", len(rows), m.selectedGroup, m.selectedServer)
	renderHomeOverviewList(overviewList, rows, m.selectedServer, m.selectedGroup, "No servers found. Press 'A' to add one.")
	scrollHomeOverviewToSelection(overviewList, rows, m.selectedServer)
}

func (m *homeScreenModel) syncGroupFromServer() {
	if m.selectedServer < 0 || m.selectedServer >= len(m.items) {
		return
	}

	group := m.items[m.selectedServer].Group
	if idx, exists := m.groupIndexByName[group]; exists {
		m.selectedGroup = idx
	}
}

func (m *homeScreenModel) syncServerFromGroup() {
	if m.selectedGroup < 0 || m.selectedGroup >= len(m.groupNames) {
		return
	}

	serverIdx, exists := m.firstServerByGroup[m.groupNames[m.selectedGroup]]
	if !exists {
		return
	}
	m.selectedServer = serverIdx
}

func (m *homeScreenModel) clampSelectedServer() {
	if len(m.items) == 0 {
		m.selectedServer = -1
		return
	}
	if m.selectedServer < 0 {
		m.selectedServer = 0
	}
	if m.selectedServer >= len(m.items) {
		m.selectedServer = len(m.items) - 1
	}
}

type homeController struct {
	model        *homeScreenModel
	overviewList *tui.Element
	action       HomeAction
}

func newHomeController(model *homeScreenModel) *homeController {
	return &homeController{
		model:  model,
		action: HomeQuit,
	}
}

func (c *homeController) attachOverviewList(overviewList *tui.Element) {
	c.overviewList = overviewList
}

func (c *homeController) render() {
	if c.overviewList == nil {
		return
	}
	c.model.render(c.overviewList)
}

func (c *homeController) selectedEntryIndex() int {
	return c.model.entryIndex()
}

func (c *homeController) handleKey(ke tui.KeyEvent) bool {
	switch ke.Key {
	case tui.KeyUp:
		c.model.moveUp()
		c.render()
		return true
	case tui.KeyDown:
		c.model.moveDown()
		c.render()
		return true
	case tui.KeyLeft:
		c.model.advanceGroup(-1)
		c.render()
		return true
	case tui.KeyRight, tui.KeyTab:
		c.model.advanceGroup(1)
		c.render()
		return true
	case tui.KeyEnter:
		return c.handleSelectionAction(ke, HomeConnect, "home connect selected index=%d", "home enter ignored: no server selected")
	case tui.KeyEscape:
		debugf("home action quit (esc)")
		c.stop(ke, HomeQuit)
		return true
	default:
		return c.handleRune(ke)
	}
}

func (c *homeController) handleRune(ke tui.KeyEvent) bool {
	r, ok := lowerRune(ke)
	if !ok {
		return false
	}

	switch r {
	case 'k':
		c.model.moveUp()
		c.render()
		return true
	case 'j':
		c.model.moveDown()
		c.render()
		return true
	case 'l':
		c.model.advanceGroup(1)
		c.render()
		return true
	case 'h', '?':
		debugf("home action help")
		c.stop(ke, HomeHelp)
		return true
	case 'a':
		debugf("home action add")
		c.stop(ke, HomeAdd)
		return true
	case 'g':
		debugf("home action add group")
		c.stop(ke, HomeAddGroup)
		return true
	case 'd':
		return c.handleSelectionAction(ke, HomeDelete, "home action delete index=%d", "home delete ignored: no server")
	case 'm':
		debugf("home action menu")
		c.stop(ke, HomeMenu)
		return true
	case 'q':
		debugf("home action quit")
		c.stop(ke, HomeQuit)
		return true
	default:
		return false
	}
}

func (c *homeController) handleSelectionAction(ke tui.KeyEvent, action HomeAction, successLog string, emptyLog string) bool {
	entryIdx := c.selectedEntryIndex()
	if entryIdx < 0 {
		debugf("%s", emptyLog)
		return true
	}
	debugf(successLog, entryIdx)
	c.stop(ke, action)
	return true
}

func (c *homeController) stop(ke tui.KeyEvent, action HomeAction) {
	c.action = action
	ke.App().Stop()
}
