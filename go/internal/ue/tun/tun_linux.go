//go:build linux

package tun

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

const (
	cloneDevicePath = "/dev/net/tun"
	// TUNSETIFF is defined in <linux/if_tun.h>
	ioctlTunSetIff = 0x400454ca
	iffTun         = 0x0001
	iffNoPi        = 0x1000
)

// Device represents a TUN device.
type Device struct {
	file *os.File
	name string
}

type ifReq struct {
	Name  [16]byte
	Flags uint16
	_     [22]byte
}

// Allocate creates a new TUN interface.
func Allocate(name string) (*Device, error) {
	file, err := os.OpenFile(cloneDevicePath, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", cloneDevicePath, err)
	}

	var req ifReq
	req.Flags = iffTun | iffNoPi
	copy(req.Name[:], name)

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), uintptr(ioctlTunSetIff), uintptr(unsafe.Pointer(&req)))
	if errno != 0 {
		file.Close()
		return nil, fmt.Errorf("ioctl TUNSETIFF failed: %w", errno)
	}

	// Extract the actual name assigned (important if name had %d)
	assignedName := string(req.Name[:])
	// trim null terminators
	for i := 0; i < len(assignedName); i++ {
		if assignedName[i] == 0 {
			assignedName = assignedName[:i]
			break
		}
	}

	return &Device{
		file: file,
		name: assignedName,
	}, nil
}

func (d *Device) Name() string {
	return d.name
}

func (d *Device) Read(p []byte) (n int, err error) {
	return syscall.Read(int(d.file.Fd()), p)
}

func (d *Device) Write(p []byte) (n int, err error) {
	return syscall.Write(int(d.file.Fd()), p)
}

func (d *Device) Close() error {
	return d.file.Close()
}

// Configure sets up the IP, netmask, MTU and optionally routing.
// Instead of complex ioctls like the C++ version, we use the `ip` command for safety and clarity.
func (d *Device) Configure(ipAddr, netmask string, mtu int, configureRoute bool) error {
	// Calculate CIDR prefix length from netmask
	// For simplicity in rewriting, we assume standard /24 or /32 or similar setups,
	// but the `ip` command accepts CIDR notation. Let's just use `ip addr add IP/MASK dev NAME`.
	// For a quick fix, let's assume ipAddr includes the CIDR or we just set it up.
	// Actually `ipconfig` allows setting mask directly via broadcast/etc but `ip addr` is standard.

	// Set MTU
	if err := runCmd("ip", "link", "set", "dev", d.name, "mtu", fmt.Sprintf("%d", mtu)); err != nil {
		return fmt.Errorf("failed to set MTU: %w", err)
	}

	// Set IP and Netmask (simplification: if netmask is not in CIDR, we should convert it,
	// but ip addr add accepts IP peer IP or just setting the address and bringing it up).
	// Let's just use ifconfig for exact match with C++ behavior or ip addr.
	// Since we know netmask like "255.255.255.0", we can convert it.
	cidr, err := netmaskToCIDR(netmask)
	if err != nil {
		return err
	}

	ipCIDR := fmt.Sprintf("%s/%d", ipAddr, cidr)
	if err := runCmd("ip", "addr", "add", ipCIDR, "dev", d.name); err != nil {
		return fmt.Errorf("failed to set IP address: %w", err)
	}

	// Bring interface up
	if err := runCmd("ip", "link", "set", "dev", d.name, "up"); err != nil {
		return fmt.Errorf("failed to bring interface up: %w", err)
	}

	if configureRoute {
		// Implement basic route configuration here.
		// The C++ version sets up specific rt_tables to avoid conflicts.
		// For the initial Go rewrite we will setup a simple direct route.
		tableName := "rt_" + d.name
		if err := setupRouting(d.name, ipAddr, tableName); err != nil {
			return fmt.Errorf("failed to configure routing: %w", err)
		}
	}

	return nil
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command %v failed: %w, output: %s", cmd.Args, err, string(out))
	}
	return nil
}

func netmaskToCIDR(netmask string) (int, error) {
	// simplistic conversion for standard IPv4 netmasks
	mask := net.ParseIP(netmask).To4()
	if mask == nil {
		return 0, fmt.Errorf("invalid netmask: %s", netmask)
	}
	cidr, _ := net.IPv4Mask(mask[0], mask[1], mask[2], mask[3]).Size()
	return cidr, nil
}

// setupRouting loosely mimics the C++ AddIpRoutes/AddNewIpRules logic.
func setupRouting(ifName, ipAddr, tableName string) error {
	// First, try to add table to /etc/iproute2/rt_tables (simplified, appending if not exists)
	// For production, this requires parsing the file cleanly.
	// As a fast approach for testing the rewrite:
	_ = runCmd("sh", "-c", fmt.Sprintf("grep -q %s /etc/iproute2/rt_tables || echo '200 %s' >> /etc/iproute2/rt_tables", tableName, tableName))

	_ = runCmd("ip", "rule", "del", "from", ipAddr, "lookup", tableName)
	_ = runCmd("ip", "rule", "add", "from", ipAddr, "table", tableName)
	_ = runCmd("ip", "route", "del", "default", "dev", ifName, "table", tableName)
	_ = runCmd("ip", "route", "add", "default", "dev", ifName, "table", tableName)

	return nil
}
