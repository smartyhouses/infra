package internal

import (
	"context"
	"fmt"
	"net"
	"runtime"

	"github.com/coreos/go-iptables/iptables"
	"github.com/txn2/txeh"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"go.opentelemetry.io/otel/trace"

	"github.com/e2b-dev/infra/packages/env-instance-task-driver/internal/slot"
	"github.com/e2b-dev/infra/packages/env-instance-task-driver/internal/telemetry"
)

const (
	hostDefaultGateway = "ens4"
	loNS               = "lo"
)

func CreateNetwork(
	ctx context.Context,
	ipSlot *slot.IPSlot,
	hosts *txeh.Hosts,
	tracer trace.Tracer,
) error {
	childCtx, childSpan := tracer.Start(ctx, "create-network")
	defer childSpan.End()

	// Prevent thread changes so the we can safely manipulate with namespaces
	telemetry.ReportEvent(childCtx, "waiting for OS thread lock")
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	telemetry.ReportEvent(childCtx, "OS thread lock passed")

	// Save the original (host) namespace and restore it upon function exit
	hostNS, err := netns.Get()
	if err != nil {
		errMsg := fmt.Errorf("cannot get current (host) namespace %w", err)
		return errMsg
	}
	telemetry.ReportEvent(childCtx, "Saved original ns")
	defer func() {
		err = netns.Set(hostNS)
		if err != nil {
			errMsg := fmt.Errorf("error resetting network namespace back to the host namespace %w", err)
			telemetry.ReportError(childCtx, errMsg)
		}
		err = hostNS.Close()
		if err != nil {
			errMsg := fmt.Errorf("error closing host network namespace %w", err)
			telemetry.ReportError(childCtx, errMsg)
		}
	}()

	// Create NS for the env instance
	ns, err := netns.NewNamed(ipSlot.NamespaceID())
	if err != nil {
		return fmt.Errorf("cannot create new namespace %w", err)
	}
	telemetry.ReportEvent(childCtx, "Created ns")
	defer ns.Close()

	// Create the Veth and Vpeer
	vethAttrs := netlink.NewLinkAttrs()
	vethAttrs.Name = ipSlot.VethName()
	veth := &netlink.Veth{
		LinkAttrs: vethAttrs,
		PeerName:  ipSlot.VpeerName(),
	}
	err = netlink.LinkAdd(veth)
	if err != nil {
		return fmt.Errorf("error creating veth device %w", err)
	}
	telemetry.ReportEvent(childCtx, "Created veth device")

	vpeer, err := netlink.LinkByName(ipSlot.VpeerName())
	if err != nil {
		return fmt.Errorf("error finding vpeer %w", err)
	}
	telemetry.ReportEvent(childCtx, "Linked veth")

	err = netlink.LinkSetUp(vpeer)
	if err != nil {
		return fmt.Errorf("error setting vpeer device up %w", err)
	}
	telemetry.ReportEvent(childCtx, "Set up veth")

	ip, ipNet, err := net.ParseCIDR(ipSlot.VpeerCIDR())
	if err != nil {
		return fmt.Errorf("error parsing vpeer CIDR %w", err)
	}
	telemetry.ReportEvent(childCtx, "Parsed CIDR")

	err = netlink.AddrAdd(vpeer, &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ip,
			Mask: ipNet.Mask,
		},
	})
	if err != nil {
		return fmt.Errorf("error adding vpeer device address %w", err)
	}
	telemetry.ReportEvent(childCtx, "Added veth address")

	// Move Veth device to the host NS
	err = netlink.LinkSetNsFd(veth, int(hostNS))
	if err != nil {
		return fmt.Errorf("error moving veth device to the host namespace %w", err)
	}
	telemetry.ReportEvent(childCtx, "Moved veth to host")

	err = netns.Set(hostNS)
	if err != nil {
		return fmt.Errorf("error setting network namespace %w", err)
	}
	telemetry.ReportEvent(childCtx, "Set ns to host ns")

	vethInHost, err := netlink.LinkByName(ipSlot.VethName())
	if err != nil {
		return fmt.Errorf("error finding veth %w", err)
	}
	telemetry.ReportEvent(childCtx, "Found veth")

	err = netlink.LinkSetUp(vethInHost)
	if err != nil {
		return fmt.Errorf("error setting veth device up %w", err)
	}
	telemetry.ReportEvent(childCtx, "Set veth device up")

	ip, ipNet, err = net.ParseCIDR(ipSlot.VethCIDR())
	if err != nil {
		return fmt.Errorf("error parsing veth  CIDR %w", err)
	}
	telemetry.ReportEvent(childCtx, "Parsed CIDR")

	err = netlink.AddrAdd(vethInHost, &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ip,
			Mask: ipNet.Mask,
		},
	})
	if err != nil {
		return fmt.Errorf("error adding veth device address %w", err)
	}
	telemetry.ReportEvent(childCtx, "Added veth device address")

	err = netns.Set(ns)
	if err != nil {
		return fmt.Errorf("error setting network namespace to %s %w", ns.String(), err)
	}
	telemetry.ReportEvent(childCtx, "Set network namespace")

	// Create Tap device for FC in NS
	tapAttrs := netlink.NewLinkAttrs()
	tapAttrs.Name = ipSlot.TapName()
	tapAttrs.Namespace = ns
	tap := &netlink.Tuntap{
		Mode:      netlink.TUNTAP_MODE_TAP,
		LinkAttrs: tapAttrs,
	}
	err = netlink.LinkAdd(tap)
	if err != nil {
		return fmt.Errorf("error creating tap device %w", err)
	}
	telemetry.ReportEvent(childCtx, "Created tap device")

	err = netlink.LinkSetUp(tap)
	if err != nil {
		return fmt.Errorf("error setting tap device up %w", err)
	}
	telemetry.ReportEvent(childCtx, "Set tap device up")

	ip, ipNet, err = net.ParseCIDR(ipSlot.TapCIDR())
	if err != nil {
		return fmt.Errorf("error parsing tap CIDR %w", err)
	}
	telemetry.ReportEvent(childCtx, "Parsed CIDR")

	err = netlink.AddrAdd(tap, &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ip,
			Mask: ipNet.Mask,
		},
	})
	if err != nil {
		return fmt.Errorf("error setting address of the tap device %w", err)
	}
	telemetry.ReportEvent(childCtx, "Set tap device address")

	// Set NS lo device up
	lo, err := netlink.LinkByName(loNS)
	if err != nil {
		return fmt.Errorf("error finding lo %w", err)
	}
	telemetry.ReportEvent(childCtx, "Found lo")

	err = netlink.LinkSetUp(lo)
	if err != nil {
		return fmt.Errorf("error setting lo device up %w", err)
	}
	telemetry.ReportEvent(childCtx, "Set lo device up")

	// Add NS default route
	err = netlink.RouteAdd(&netlink.Route{
		Scope: netlink.SCOPE_UNIVERSE,
		Gw:    net.ParseIP(ipSlot.VethIP()),
	})
	if err != nil {
		return fmt.Errorf("error adding default NS route %w", err)
	}
	telemetry.ReportEvent(childCtx, "Added default ns route")

	tables, err := iptables.New()
	if err != nil {
		return fmt.Errorf("error initializing iptables %w", err)
	}
	telemetry.ReportEvent(childCtx, "Initialized iptables")

	// Add NAT routing rules to NS
	err = tables.Append("nat", "POSTROUTING", "-o", ipSlot.VpeerName(), "-s", ipSlot.NamespaceSnapshotIP(), "-j", "SNAT", "--to", ipSlot.HostSnapshotIP())
	if err != nil {
		errMsg := fmt.Errorf("error creating postrouting rule to vpeer %w", err)
		return errMsg
	}
	telemetry.ReportEvent(childCtx, "Created postrouting rule to vpeer")

	err = tables.Append("nat", "PREROUTING", "-i", ipSlot.VpeerName(), "-d", ipSlot.HostSnapshotIP(), "-j", "DNAT", "--to", ipSlot.NamespaceSnapshotIP())
	if err != nil {
		errMsg := fmt.Errorf("error creating postrouting rule from vpeer %w", err)
		return errMsg
	}
	telemetry.ReportEvent(childCtx, "Created postrouting rule from vpeer")

	// Go back to original namespace
	err = netns.Set(hostNS)
	if err != nil {
		return fmt.Errorf("error setting network namespace to %s %w", hostNS.String(), err)
	}
	telemetry.ReportEvent(childCtx, "Set network namespace back")

	// Add routing from host to FC namespace
	_, ipNet, err = net.ParseCIDR(ipSlot.HostSnapshotCIDR())
	if err != nil {
		return fmt.Errorf("error parsing host snapshot CIDR %w", err)
	}
	telemetry.ReportEvent(childCtx, "Parsed CIDR")

	err = netlink.RouteAdd(&netlink.Route{
		Gw:  net.ParseIP(ipSlot.VpeerIP()),
		Dst: ipNet,
	})
	if err != nil {
		return fmt.Errorf("error adding route from host to FC %w", err)
	}
	telemetry.ReportEvent(childCtx, "Added route from host to FC")

	// Add host forwarding rules
	err = tables.Append("filter", "FORWARD", "-i", ipSlot.VethName(), "-o", hostDefaultGateway, "-j", "ACCEPT")
	if err != nil {
		errMsg := fmt.Errorf("error creating forwarding rule to default gateway %w", err)
		return errMsg
	}
	telemetry.ReportEvent(childCtx, "Created forwarding rule to default gateway")

	err = tables.Append("filter", "FORWARD", "-i", hostDefaultGateway, "-o", ipSlot.VethName(), "-j", "ACCEPT")
	if err != nil {
		errMsg := fmt.Errorf("error creating forwarding rule from default gateway %w", err)
		return errMsg
	}
	telemetry.ReportEvent(childCtx, "Created forwarding rule from default gateway")

	// Add host postrouting rules
	err = tables.Append("nat", "POSTROUTING", "-s", ipSlot.HostSnapshotCIDR(), "-o", hostDefaultGateway, "-j", "MASQUERADE")
	if err != nil {
		errMsg := fmt.Errorf("error creating postrouting rule %w", err)
		return errMsg
	}
	telemetry.ReportEvent(childCtx, "Created postrouting rule")

	// Add entry to etc hosts
	hosts.AddHost(ipSlot.HostSnapshotIP(), ipSlot.InstanceID)
	err = hosts.Save()
	if err != nil {
		return fmt.Errorf("error adding env instance to etc hosts %w", err)
	}
	telemetry.ReportEvent(childCtx, "Added env instance to etc hosts")

	return nil
}

func RemoveNetwork(ctx context.Context, ipSlot *slot.IPSlot, hosts *txeh.Hosts, consulToken string, tracer trace.Tracer) error {
	childCtx, childSpan := tracer.Start(ctx, "remove-network")
	defer childSpan.End()

	hosts.RemoveHost(ipSlot.InstanceID)
	err := hosts.Save()
	if err != nil {
		errMsg := fmt.Errorf("error removing env instance to etc hosts %w", err)
		telemetry.ReportCriticalError(childCtx, errMsg)
	} else {
		telemetry.ReportEvent(childCtx, "removed env instance to etc hosts")
	}

	tables, err := iptables.New()
	if err != nil {
		errMsg := fmt.Errorf("error initializing iptables %w", err)
		telemetry.ReportCriticalError(childCtx, errMsg)
	} else {
		telemetry.ReportEvent(childCtx, "initialized iptables")

		// Delete host forwarding rules
		err = tables.Delete("filter", "FORWARD", "-i", ipSlot.VethName(), "-o", hostDefaultGateway, "-j", "ACCEPT")
		if err != nil {
			errMsg := fmt.Errorf("error deleting host forwarding rule to default gateway %w", err)
			telemetry.ReportCriticalError(childCtx, errMsg)
		} else {
			telemetry.ReportEvent(childCtx, "deleted host forwarding rule to default gateway")
		}

		err = tables.Delete("filter", "FORWARD", "-i", hostDefaultGateway, "-o", ipSlot.VethName(), "-j", "ACCEPT")
		if err != nil {
			errMsg := fmt.Errorf("error deleting host forwarding rule from default gateway %w", err)
			telemetry.ReportCriticalError(childCtx, errMsg)
		} else {
			telemetry.ReportEvent(childCtx, "deleted host forwarding rule from default gateway")
		}

		// Delete host postrouting rules
		err = tables.Delete("nat", "POSTROUTING", "-s", ipSlot.HostSnapshotCIDR(), "-o", hostDefaultGateway, "-j", "MASQUERADE")
		if err != nil {
			errMsg := fmt.Errorf("error deleting host postrouting rule %w", err)
			telemetry.ReportCriticalError(childCtx, errMsg)
		} else {
			telemetry.ReportEvent(childCtx, "deleted host postrouting rule")
		}
	}

	// Delete routing from host to FC namespace
	_, ipNet, err := net.ParseCIDR(ipSlot.HostSnapshotCIDR())
	if err != nil {
		errMsg := fmt.Errorf("error parsing host snapshot CIDR %w", err)
		telemetry.ReportCriticalError(childCtx, errMsg)
	} else {
		telemetry.ReportEvent(childCtx, "parsed CIDR")

		err = netlink.RouteDel(&netlink.Route{
			Gw:  net.ParseIP(ipSlot.VpeerIP()),
			Dst: ipNet,
		})
		if err != nil {
			errMsg := fmt.Errorf("error deleting route from host to FC %w", err)
			telemetry.ReportCriticalError(childCtx, errMsg)
		} else {
			telemetry.ReportEvent(childCtx, "deleted route from host to FC")
		}
	}

	// Delete veth device
	// We explicitly delete the veth device from the host namespace because even though deleting
	// is deleting the device there may be a race condition when creating a new veth device with
	// the same name immediately after deleting the namespace.
	veth, err := netlink.LinkByName(ipSlot.VethName())
	if err != nil {
		errMsg := fmt.Errorf("error finding veth %w", err)
		telemetry.ReportCriticalError(childCtx, errMsg)
	} else {
		telemetry.ReportEvent(childCtx, "found veth")

		err = netlink.LinkDel(veth)
		if err != nil {
			errMsg := fmt.Errorf("error deleting veth device %w", err)
			telemetry.ReportCriticalError(childCtx, errMsg)
		} else {
			telemetry.ReportEvent(childCtx, "deleted veth device")
		}
	}

	err = netns.DeleteNamed(ipSlot.NamespaceID())
	if err != nil {
		errMsg := fmt.Errorf("error deleting namespace %w", err)
		telemetry.ReportCriticalError(childCtx, errMsg)
	} else {
		telemetry.ReportEvent(childCtx, "deleted namespace")
	}

	err = ipSlot.Release(childCtx, consulToken, tracer)
	if err != nil {
		errMsg := fmt.Errorf("error releasing slot %w", err)
		telemetry.ReportCriticalError(childCtx, errMsg)
		return errMsg
	}

	telemetry.ReportEvent(childCtx, "released slot")

	return nil
}
