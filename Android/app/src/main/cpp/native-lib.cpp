#include <jni.h>
#include <string>
#include "libfoo.h"

extern "C"
void
Java_tk_cocoon_MainActivity_serverJNI() {
    server();
}