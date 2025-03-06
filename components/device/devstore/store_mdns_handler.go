package devstore

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/open-control-systems/device-hub/components/status"
	"github.com/open-control-systems/device-hub/components/system/sysmdns"
)

// StoreMdnsHandler notifies store about new devices discovered over local network.
type StoreMdnsHandler struct {
	store Store
}

// NewStoreMdnsHandler is an initialization of StoreMdnsHandler.
//
// Parameters:
//   - store to automatically add devices discovered in the local network.
func NewStoreMdnsHandler(store Store) *StoreMdnsHandler {
	return &StoreMdnsHandler{store: store}
}

// HandleService handles mDNS service discovered over local network.
func (h *StoreMdnsHandler) HandleService(service *sysmdns.Service) error {
	if ignoreService(service) {
		return nil
	}

	records, err := parseTxtRecords(service.TxtRecords)
	if err != nil {
		return err
	}

	modeStr, ok := records["autodiscovery_mode"]
	if !ok {
		return nil
	}

	mode, err := parseAutodiscoveryMode(modeStr)
	if err != nil {
		return err
	}

	uri, ok := records["autodiscovery_uri"]
	if !ok {
		return nil
	}

	desc, ok := records["autodiscovery_desc"]
	if !ok {
		return nil
	}

	typ, ok := records["autodiscovery_type"]
	if !ok {
		return nil
	}

	return h.handleAutodiscovery(mode, uri, typ, desc)
}

func (h *StoreMdnsHandler) handleAutodiscovery(
	mode autodiscoveryMode,
	uri string,
	typ string,
	desc string,
) error {
	switch mode {
	case autodiscoveryModeAdd:
		return h.handleAutodiscoveryAdd(uri, typ, desc)
	default:
		panic("failed to handle device auto-discovery: invalid state")
	}
}

func (h *StoreMdnsHandler) handleAutodiscoveryAdd(uri string, typ string, desc string) error {
	err := h.store.Add(uri, typ, desc)
	if err != nil && err != ErrDeviceExist {
		return err
	}

	return nil
}

type autodiscoveryMode int

const (
	autodiscoveryModeInvalid autodiscoveryMode = iota
	autodiscoveryModeAdd
)

func ignoreService(service *sysmdns.Service) bool {
	for _, record := range service.TxtRecords {
		if strings.Contains(record, "autodiscovery_mode") {
			return false
		}
	}

	return true
}

func parseTxtRecords(records []string) (map[string]string, error) {
	ret := make(map[string]string)

	for _, record := range records {
		if !strings.Contains(record, "=") {
			continue
		}

		tokens := strings.Split(record, "=")
		if len(tokens) != 2 {
			return nil, status.StatusInvalidArg
		}

		if tokens[0] == "" || tokens[1] == "" {
			return nil, status.StatusInvalidArg
		}

		ret[tokens[0]] = tokens[1]
	}

	return ret, nil
}

func parseAutodiscoveryMode(str string) (autodiscoveryMode, error) {
	mode, err := strconv.Atoi(str)
	if err != nil {
		return autodiscoveryModeInvalid,
			fmt.Errorf("failed to parse mdns_autodiscovery: mode=%s err=%v", str, err)
	}

	switch mode {
	case 1:
		return autodiscoveryModeAdd, nil
	default:
	}

	return autodiscoveryModeInvalid, status.StatusInvalidArg
}
