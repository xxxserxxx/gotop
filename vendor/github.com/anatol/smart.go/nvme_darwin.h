// SPDX-License-Identifier: MIT

#ifdef __cplusplus
extern "C" {
#endif

unsigned int smart_nvme_open_darwin(const char *path, void **ptr);
unsigned int smart_nvme_identify_darwin(void *ptr, void *buffer, unsigned int nsid);
unsigned int smart_nvme_readsmart_darwin(void *ptr, void *buffer);
void smart_nvme_close_darwin(void *ptr);

#ifdef __cplusplus
}
#endif