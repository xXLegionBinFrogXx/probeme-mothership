#include "shim.h"

#include <dlfcn.h>
#include <stddef.h>

void *pme_shim_open(const char *path, const char **err) {
    *err = NULL;
    void *h = dlopen(path, RTLD_NOW | RTLD_LOCAL);
    if (h == NULL) {
        *err = dlerror();
    }
    return h;
}

const struct pme_provider *pme_shim_get(void *handle, const char **err) {
    *err = NULL;
    dlerror(); /* clear any stale error */
    void *sym = dlsym(handle, "pme_provider_get");
    const char *d = dlerror();
    if (d != NULL || sym == NULL) {
        if (d == NULL) {
            d = "symbol is NULL";
        }
        *err = d;
        return NULL;
    }
    const struct pme_provider *(*get)(void) =
        (const struct pme_provider *(*)(void))sym;
    return get();
}

void pme_shim_close(void *handle) {
    if (handle != NULL) {
        dlclose(handle);
    }
}
