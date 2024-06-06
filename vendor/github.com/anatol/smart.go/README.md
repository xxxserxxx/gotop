# smart.go

Smart.go is a pure Golang library to access disk low-level [S.M.A.R.T.](https://en.wikipedia.org/wiki/S.M.A.R.T.) information.
Smart.go tries to match functionality provided by [smartctl](https://www.smartmontools.org/) but with golang API.

Currently this library support SATA, SCSI and NVMe drives. Different drive types provide different set of monitoring information and API reflects it.

At this point the library works at Linux and partially at MacOSX. We are looking for help with porting it to other platforms.

## Example

Here is an example of code that demonstrates the library usage.

```go
// skip the error handling for more compact API example
dev, _ := smart.OpenNVMe("/dev/nvme0n1")
c, nss, _ := dev.Identify()
fmt.Println("Model number: ", c.ModelNumber())
fmt.Println("Serial number: ", c.SerialNumber())
fmt.Println("Size: ", c.Tnvmcap.Val[0])

// namespace #1
ns := nss[0]
fmt.Println("Namespace 1 utilization: ", ns.Nuse*ns.LbaSize())

sm, _ := dev.ReadSMART()
fmt.Println("Temperature: ", sm.Temperature, "K")
// PowerOnHours is reported as 128-bit value and represented by this library as an array of uint64
fmt.Println("Power-on hours: ", sm.PowerOnHours.Val[0])
fmt.Println("Power cycles: ", sm.PowerCycles.Val[0])
```

The output looks like
```text
Model number:  SAMSUNG MZVLB512HBJQ-000L7
Serial number:  S4ENNF0M741521
Size:  512110190592
Namespace 1 utilization:  387524902912
Temperature:  327 K
Power-on hours:  499
Power cycles:  1433
```

Here is an example of iterating over system's block devices:
```go
block, err := ghw.Block()
if err != nil {
  panic(err)
}
for _, disk := range block.Disks {
        dev, err := smart.Open("/dev/" + disk.Name)
        if err != nil {
            // some devices (like dmcrypt) do not support SMART interface
            fmt.Println(err)
            continue
        }
        defer dev.Close()

        switch sm := dev.(type) {
        case *smart.SataDevice:
            data, err := sm.ReadSMARTData()
            attr, ok := data.Attrs[194]; ok { // attr.Name == "Temperature_Celsius"
                temp, min, max, overtempCounter, err := attr.ParseAsTemperature()
                // min/max/counter are optional
            }
        case *smart.ScsiDevice:
            _, _ = sm.Capacity()
        case *smart.NVMeDevice:
            _, _ = sm.ReadSMART()
        }
}
```

Reading generic SMART attributes.

smart.go provides API for easier access to the most commonly used device attributes.

```go
dev, err := smart.Open("/dev/nvme0n1")
require.NoError(t, err)
defer dev.Close()

a, err := dev.ReadGenericAttributes()
require.NoError(t, err)

fmt.Println("The temperature is ", a.Temperature) // in Celsius
fmt.Println("Read block count ", a.Read)
fmt.Println("Written block count ", a.Written)
fmt.Println("Power Cycles count ", a.PowerCycles)
fmt.Println("Power On Hours ", a.PowerOnHours)
```

### Credit
This project is inspired by https://github.com/dswarbrick/smart
