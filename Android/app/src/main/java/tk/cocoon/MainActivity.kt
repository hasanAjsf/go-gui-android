package tk.cocoon

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