// Contains code with this licence:
// Copyright © 2019 Jonathan Pentecost <pentecostjonathan@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chromecastpredicate

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/vishen/go-chromecast/application"
	castdns "github.com/vishen/go-chromecast/dns"
	"github.com/vishen/go-chromecast/storage"
)

var (
	cache = storage.NewStorage()

	// Set up a global dns entry so we can attempt reconnects
	entry castdns.CastDNSEntry
)

type cachedDNSEntry struct {
	UUID string
	Name string
	Addr string
	Port int
}

// GetUUID returns UUID
func (e cachedDNSEntry) GetUUID() string {
	return e.UUID
}

// GetName returns Name
func (e cachedDNSEntry) GetName() string {
	return e.Name
}

// GetAddr returns Addr
func (e cachedDNSEntry) GetAddr() string {
	return e.Addr
}

// GetPort returns Port
func (e cachedDNSEntry) GetPort() int {
	return e.Port
}

func castApplication(addr string) (*application.Application, error) {
	// TODO move to predicate parameters
	deviceName := ""
	disableCache := false
	port := 8009
	ifaceName := ""
	dnsTimeoutSeconds := 3
	device := ""
	useFirstDevice := true
	deviceUUID := ""
	debug := true

	// Used to try and reconnect
	if deviceUUID == "" && entry != nil {
		deviceUUID = entry.GetUUID()
		entry = nil
	}

	applicationOptions := []application.ApplicationOption{
		application.WithDebug(debug),
		application.WithCacheDisabled(disableCache),
	}

	// If we need to look on a specific network interface for mdns or
	// for finding a network ip to host from, ensure that the network
	// interface exists.
	var iface *net.Interface
	if ifaceName != "" {
		var err error
		if iface, err = net.InterfaceByName(ifaceName); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("unable to find interface %q", ifaceName))
		}
		applicationOptions = append(applicationOptions, application.WithIface(iface))
	}

	// If no address was specified, attempt to determine the address of any
	// local chromecast devices.
	if addr == "" {
		// If a device name or uuid was specified, check the cache for the ip+port
		found := false
		if !disableCache && (deviceName != "" || deviceUUID != "") {
			entry = findCachedCastDNS(deviceName, deviceUUID)
			found = entry.GetAddr() != ""
		}
		if !found {
			var err error
			if entry, err = findCastDNS(iface, dnsTimeoutSeconds, device, deviceName, deviceUUID, useFirstDevice); err != nil {
				return nil, errors.Wrap(err, "unable to find cast dns entry")
			}
		}
		if !disableCache {
			cachedEntry := cachedDNSEntry{
				UUID: entry.GetUUID(),
				Name: entry.GetName(),
				Addr: entry.GetAddr(),
				Port: entry.GetPort(),
			}
			cachedEntryJSON, _ := json.Marshal(cachedEntry)
			if err := cache.Save(getCacheKey(cachedEntry.UUID), cachedEntryJSON); err != nil {
				return nil, err
			}
			if err := cache.Save(getCacheKey(cachedEntry.Name), cachedEntryJSON); err != nil {
				return nil, err
			}
		}
		if debug {
			fmt.Printf("using device name=%s addr=%s port=%d uuid=%s\n", entry.GetName(), entry.GetAddr(), entry.GetPort(), entry.GetUUID())
		}
	} else {
		p := port
		entry = cachedDNSEntry{
			Addr: addr,
			Port: p,
		}
	}
	app := application.NewApplication(applicationOptions...)
	if err := app.Start(entry.GetAddr(), entry.GetPort()); err != nil {
		// NOTE: currently we delete the dns cache every time we get
		// an error, this is to make sure that if the device gets a new
		// ipaddress we will invalidate the cache.
		if err := cache.Save(getCacheKey(entry.GetUUID()), []byte{}); err != nil {
			return nil, err
		}
		if err := cache.Save(getCacheKey(entry.GetName()), []byte{}); err != nil {
			return nil, err
		}
		return nil, err
	}
	return app, nil
}

func getCacheKey(suffix string) string {
	return fmt.Sprintf("cmd/utils/dns/%s", suffix)
}

func findCachedCastDNS(deviceName, deviceUUID string) castdns.CastDNSEntry {
	for _, s := range []string{deviceName, deviceUUID} {
		cacheKey := getCacheKey(s)
		if b, err := cache.Load(cacheKey); err == nil {
			cachedEntry := cachedDNSEntry{}
			if err := json.Unmarshal(b, &cachedEntry); err == nil {
				return cachedEntry
			}
		}
	}
	return cachedDNSEntry{}
}

func findCastDNS(iface *net.Interface, dnsTimeoutSeconds int, device, deviceName, deviceUUID string, first bool) (castdns.CastDNSEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(dnsTimeoutSeconds))
	defer cancel()
	castEntryChan, err := castdns.DiscoverCastDNSEntries(ctx, iface)
	if err != nil {
		return castdns.CastEntry{}, err
	}

	foundEntries := []castdns.CastEntry{}
	for entry := range castEntryChan {
		if first || (deviceUUID != "" && entry.UUID == deviceUUID) || (deviceName != "" && entry.DeviceName == deviceName) || (device != "" && entry.Device == device) {
			return entry, nil
		}
		foundEntries = append(foundEntries, entry)
	}

	if len(foundEntries) == 0 {
		return castdns.CastEntry{}, fmt.Errorf("no cast devices found on network")
	}

	// Always return entries in deterministic order.
	sort.Slice(foundEntries, func(i, j int) bool { return foundEntries[i].DeviceName < foundEntries[j].DeviceName })

	fmt.Printf("Found %d cast dns entries, select one:\n", len(foundEntries))
	for i, d := range foundEntries {
		fmt.Printf("%d) device=%q device_name=%q address=\"%s:%d\" uuid=%q\n", i+1, d.Device, d.DeviceName, d.AddrV4, d.Port, d.UUID)
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Enter selection: ")
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error reading console: %v\n", err)
			continue
		}
		i, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil {
			continue
		} else if i < 1 || i > len(foundEntries) {
			continue
		}
		return foundEntries[i-1], nil
	}
}
