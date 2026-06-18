package com.glootea.goquiz.feature.auth.data

import android.content.Context
import eu.anifantakis.lib.ksafe.KSafe

actual object KSafeFactory {
    private var appContext: Context? = null

    fun setAndroidContext(context: Context) {
        appContext = context.applicationContext
    }

    actual fun create(): KSafe {
        val ctx = appContext
            ?: error("KSafeFactory: Android context not set. Call KSafeFactory.setAndroidContext(...) from Application.onCreate.")
        return KSafe(ctx, fileName = "goquiz_auth", config = defaultConfig())
    }
}
