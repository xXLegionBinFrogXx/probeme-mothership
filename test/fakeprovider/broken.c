/* Compiles to libprobeme_broken.so: a valid .so with no pme_provider_get
 * symbol, for the missing-symbol test path. */

int not_a_provider(void) { return 42; }
