#ifndef PME_SHIM_H
#define PME_SHIM_H

/* Tiny dlopen wrappers: the only C in the repo besides the cgo preamble. */

void *pme_shim_open(const char *path, const char **err);

const struct pme_provider *pme_shim_get(void *handle, const char **err);

void pme_shim_close(void *handle);

#endif /* PME_SHIM_H */
