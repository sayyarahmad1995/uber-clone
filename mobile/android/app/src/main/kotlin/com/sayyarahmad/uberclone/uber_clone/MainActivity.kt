package com.sayyarahmad.uberclone.uber_clone

import android.content.ActivityNotFoundException
import android.content.Intent
import android.net.Uri
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel

class MainActivity : FlutterActivity() {
    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        MethodChannel(flutterEngine.dartExecutor.binaryMessenger, "higo/navigation")
            .setMethodCallHandler { call, result ->
                if (call.method != "openGoogleMaps") {
                    result.notImplemented()
                    return@setMethodCallHandler
                }
                val latitude = (call.argument<Any>("latitude") as? Number)?.toDouble()
                val longitude = (call.argument<Any>("longitude") as? Number)?.toDouble()
                if (latitude == null || longitude == null ||
                    !latitude.isFinite() || !longitude.isFinite() ||
                    latitude !in -90.0..90.0 || longitude !in -180.0..180.0
                ) {
                    result.success(false)
                    return@setMethodCallHandler
                }
                // Omit origin: Maps uses its current device location for navigation.
                val uri = Uri.Builder()
                    .scheme("https")
                    .authority("www.google.com")
                    .path("/maps/dir/")
                    .appendQueryParameter("api", "1")
                    .appendQueryParameter("destination", "$latitude,$longitude")
                    .appendQueryParameter("travelmode", "driving")
                    .appendQueryParameter("dir_action", "navigate")
                    .build()
                val maps = Intent(Intent.ACTION_VIEW, uri)
                    .setPackage("com.google.android.apps.maps")
                val opened = launch(maps) || launch(Intent(Intent.ACTION_VIEW, uri))
                result.success(opened)
            }
    }

    private fun launch(intent: Intent): Boolean {
        return try {
            startActivity(intent)
            true
        } catch (_: ActivityNotFoundException) {
            false
        } catch (_: SecurityException) {
            false
        }
    }
}
