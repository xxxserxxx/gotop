// SPDX-License-Identifier: MIT

#include "nvme_darwin.h"
#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/IOCFPlugIn.h>
#include <IOKit/IOKitLib.h>
#include <IOKit/storage/nvme/NVMeSMARTLibExternal.h>
#include <stdlib.h>

struct smart_nvme_darwin {
  IONVMeSMARTInterface **smartIfNVMe;
  IOCFPlugInInterface **plugin;
  io_object_t disk;
};

static bool is_smart_capable(io_object_t dev) {
  CFTypeRef smartCapableKey = IORegistryEntryCreateCFProperty(dev, CFSTR(kIOPropertyNVMeSMARTCapableKey), kCFAllocatorDefault, 0);
  if (smartCapableKey) {
    CFRelease(smartCapableKey);
    return true;
  }

  return false;
}

// *path is a string that has a format like "disk0".
unsigned int smart_nvme_open_darwin(const char *path, void **ptr) {
  // see also https://gist.github.com/AlanQuatermain/250538
  SInt32 score = 0;
  int res = EINVAL;
  IONVMeSMARTInterface **smartIfNVMe;
  IOCFPlugInInterface **plugin;
  struct smart_nvme_darwin *nvme;

  CFMutableDictionaryRef matcher = IOBSDNameMatching(kIOMainPortDefault, 0, path);
  io_object_t disk = IOServiceGetMatchingService(kIOMainPortDefault, matcher);

  while (!is_smart_capable(disk)) {
    io_object_t prevdisk = disk;

    // Find this device's parent and try again.
    IOReturn err = IORegistryEntryGetParentEntry(disk, kIOServicePlane, &disk);
    if (err != kIOReturnSuccess || !disk) {
      IOObjectRelease(prevdisk);
      break;
    }
  }

  if (!disk) {
    printf("no disk found");
    goto exit1;
  }

  res = IOCreatePlugInInterfaceForService(disk, kIONVMeSMARTUserClientTypeID, kIOCFPlugInInterfaceID, &plugin, &score);
  if (res != kIOReturnSuccess)
    goto exit2;

  res = (*plugin)->QueryInterface(plugin, CFUUIDGetUUIDBytes(kIONVMeSMARTInterfaceID), (void **)&smartIfNVMe);
  if (res != S_OK)
    goto exit3;

  *ptr = malloc(sizeof(struct smart_nvme_darwin));
  nvme = (struct smart_nvme_darwin *)*ptr;

  if (!nvme)
    goto exit4;

  nvme->disk = disk;
  nvme->plugin = plugin;
  nvme->smartIfNVMe = smartIfNVMe;

  return 0;

exit4:
  (*smartIfNVMe)->Release(smartIfNVMe);
exit3:
  IODestroyPlugInInterface(plugin);
exit2:
  IOObjectRelease(disk);
exit1:
  return res;
}

unsigned int smart_nvme_identify_darwin(void *ptr, void *buffer, unsigned int nsid) {
  struct smart_nvme_darwin *nvme = (struct smart_nvme_darwin *)ptr;
  return (*nvme->smartIfNVMe)->GetIdentifyData(nvme->smartIfNVMe, buffer, nsid);
}

unsigned int smart_nvme_readsmart_darwin(void *ptr, void *buffer) {
  struct smart_nvme_darwin *nvme = (struct smart_nvme_darwin *)ptr;
  return (*nvme->smartIfNVMe)->SMARTReadData(nvme->smartIfNVMe, (struct NVMeSMARTData *)buffer);
}

void smart_nvme_close_darwin(void *ptr) {
  struct smart_nvme_darwin *nvme = (struct smart_nvme_darwin *)ptr;
  (*nvme->smartIfNVMe)->Release(nvme->smartIfNVMe);
  IODestroyPlugInInterface(nvme->plugin);
  IOObjectRelease(nvme->disk);

  free(nvme);
}
