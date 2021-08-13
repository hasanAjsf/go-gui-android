**Problem**

I've the below simple go server that is running at my laptop (Mac/Windows/Linux):
```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there %s!", r.URL.Path[1:])
}

func main() {
	http.HandleFunc("/", handler)
	log.Println(http.ListenAndServe("localhost:6060", nil))
}
```

![Screen Shot 2021-08-13 at 3.19.24 PM|690x333](upload://vOtUW0jNTRqQ2M7mD0EVtmhQael.png) 

Can I use the same codebase to run my app at mobile `webview`, without using gomobile or other packages, so I've my code as universal app?

**Solution**

The answer is "Yes", but some slight modification to the file itself is required.

1. Remove everything from the `func main() {}` as we'll build the final result as a shared library, not as an executable binary.
2. Run the server in an `//export` function.
3. Run the server from an `anonymous goroutine` as `go func() {}()` so it is not blocking the main thread of the mobile app.
4. To keep the server gorotine running, we need to use a chanel as `<-c` to prevent the gorotine from exit.
5. Use `cgo` by adding `import "C"`, so the main file become like this:
```go
package main

import "C"

// other imports should be seperate from the special Cgo import
import (
	"fmt"
	"log"
	"net/http"
)

//export server
func server() {
	c := make(chan bool)
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
		<-c
	}()

	http.HandleFunc("/", handler)

}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there %s!", r.URL.Path[1:])
}

func main() {}
```
6. Ensure to have Androd `NDK` installed, and you know its bath
7. Build the `c-shared` output with an output name as `libxxx`, to build for `Android` use:
```bash
	CGO_ENABLED=1 \
	GOOS=android \
	GOARCH=arm \
	GOARM=7 \
	CC=$(NDK_BIN)/armv7a-linux-androideabi21-clang \
	go build -buildmode=c-shared -o libfoo.so http.go
```
**Wait** As android has multiple architectures, we need to compile for each one individually, so we can have all the process automated in a `Makefile` as below **after** creating the android app by selecting `Native C++` from the project templates, below the output library name is `libfoo` and 2 files will be generated in each folder `libfoo.so` and `libfoo.h`:

[![enter image description here][1]][1]

```make
#Filename: Makefile
# To compile run:
# make android

IOS_OUT=lib/ios
ANDROID_OUT=../android_app/app/src/main/jniLibs
ANDROID_SDK=$(HOME)/Library/Android/sdk
NDK_BIN=$(ANDROID_SDK)/ndk/23.0.7599858/toolchains/llvm/prebuilt/darwin-x86_64/bin

android-armv7a:
	CGO_ENABLED=1 \
	GOOS=android \
	GOARCH=arm \
	GOARM=7 \
	CC=$(NDK_BIN)/armv7a-linux-androideabi21-clang \
	go build -buildmode=c-shared -o $(ANDROID_OUT)/armeabi-v7a/libfoo.so ./cmd/libfoo

android-arm64:
	CGO_ENABLED=1 \
	GOOS=android \
	GOARCH=arm64 \
	CC=$(NDK_BIN)/aarch64-linux-android21-clang \
	go build -buildmode=c-shared -o $(ANDROID_OUT)/arm64-v8a/libfoo.so ./cmd/libfoo

android-x86:
	CGO_ENABLED=1 \
	GOOS=android \
	GOARCH=386 \
	CC=$(NDK_BIN)/i686-linux-android21-clang \
	go build -buildmode=c-shared -o $(ANDROID_OUT)/x86/libfoo.so ./cmd/libfoo

android-x86_64:
	CGO_ENABLED=1 \
	GOOS=android \
	GOARCH=amd64 \
	CC=$(NDK_BIN)/x86_64-linux-android21-clang \
	go build -buildmode=c-shared -o $(ANDROID_OUT)/x86_64/libfoo.so ./cmd/libfoo

android: android-armv7a android-arm64 android-x86 android-x86_64
```
8. Go to `android_app/app/src/main/cpp` and do the following:
8.1. File `CMakeLists.txt`, make it as:
```txt
cmake_minimum_required(VERSION 3.10.2)

project("android")

add_library( # Sets the name of the library.
             native-lib

             # Sets the library as a shared library.
             SHARED

             # Provides a relative path to your source file(s).
             native-lib.cpp )

add_library(lib_foo SHARED IMPORTED)
set_property(TARGET lib_foo PROPERTY IMPORTED_NO_SONAME 1)
set_target_properties(lib_foo PROPERTIES IMPORTED_LOCATION ${CMAKE_CURRENT_SOURCE_DIR}/../jniLibs/${CMAKE_ANDROID_ARCH_ABI}/libfoo.so)
include_directories(${CMAKE_CURRENT_SOURCE_DIR}/../jniLibs/${CMAKE_ANDROID_ARCH_ABI}/)

find_library( # Sets the name of the path variable.
              log-lib

              # Specifies the name of the NDK library that
              # you want CMake to locate.
              log )

target_link_libraries( # Specifies the target library.
                       native-lib
                       lib_foo

                       # Links the target library to the log library
                       # included in the NDK.
                       ${log-lib} )
```
8.2. File `native-lib.cpp` make it as:
```cpp
#include <jni.h>
#include <string>

#include "libfoo.h" // our library header

extern "C" {
    void
    Java_tk_android_MainActivity_serverJNI() {
        // Running the server
        server();
    }
}
```
9. Add webview to the `layout/activity_main`, as:
```xml
<?xml version="1.0" encoding="utf-8"?>
<androidx.constraintlayout.widget.ConstraintLayout xmlns:android="http://schemas.android.com/apk/res/android"
    xmlns:app="http://schemas.android.com/apk/res-auto"
    xmlns:tools="http://schemas.android.com/tools"
    android:layout_width="match_parent"
    android:layout_height="match_parent"
    tools:context=".MainActivity">

    <WebView
        android:id="@+id/wv"
        android:layout_width="match_parent"
        android:layout_height="match_parent"
        android:isScrollContainer="false"
        app:layout_constraintBottom_toBottomOf="parent"
        app:layout_constraintHorizontal_bias="0.0"
        app:layout_constraintLeft_toLeftOf="parent"
        app:layout_constraintRight_toRightOf="parent" />

</androidx.constraintlayout.widget.ConstraintLayout>
```
10. Update the `MainActivity` as below:
```kotlin
package tk.android

import android.os.Bundle
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.appcompat.app.AppCompatActivity

class MainActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        var wv = findViewById<WebView>(R.id.web_view)
        serverJNI()
        wv.loadUrl("http://127.0.0.1:6060/")
        wv.webViewClient = object : WebViewClient() {
            override fun shouldOverrideUrlLoading(viewx: WebView, urlx: String): Boolean {
                viewx.loadUrl(urlx)
                return false
            }
        }
    }

    private external fun serverJNI(): Void

    companion object {
        // Used to load the 'native-lib' library on application startup.
        init {
            System.loadLibrary("native-lib")
        }
    }
}
```
11. Update `AndroidManifest` to be:
```xml
<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="tk.android">

    <!-- Mandatory:
                android:usesCleartextTraffic="true"
         Optional: 
                android:hardwareAccelerated="true" 
         Depending on the action bar required:
                android:theme="@style/Theme.AppCompat.NoActionBar"
    -->
    <application
        android:hardwareAccelerated="true"     // <- Optional 
        android:usesCleartextTraffic="true"     // <- A must to be added
        android:allowBackup="true"
        android:icon="@mipmap/ic_launcher"
        android:label="@string/app_name"
        android:roundIcon="@mipmap/ic_launcher_round"
        android:supportsRtl="true"
        android:theme="@style/Theme.AppCompat.NoActionBar">   // <- If do not want action bar
        <activity android:name=".MainActivity">
            <intent-filter>
                <action android:name="android.intent.action.MAIN" />

                <category android:name="android.intent.category.LAUNCHER" />
            </intent-filter>
        </activity>
    </application>

</manifest>
```

[![enter image description here][2]][2]

**Bonus**

With Go `embed` all static files can be embed in the same library, including `css`, `javascript`, `templates` so you can buid either API, or full app with GUI

![Screen Shot 2021-08-13 at 6.35.37 PM|690x444](upload://nWOx7xTO9GxML2oFJ8eDSS5e5NA.png) 

I uploaded the main file [here](https://github.com/hajsf/go-gui-android) if any one interested about the topic.

Credit goes to [Roger Chapman](https://rogchap.com/2020/09/14/running-go-code-on-ios-and-android/)

  [1]: https://i.stack.imgur.com/heP1f.png
  [2]: https://i.stack.imgur.com/yaNbX.png

  [1]: https://i.stack.imgur.com/PAVIy.png
