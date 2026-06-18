package com.glootea.goquiz.feature.auth.data

import eu.anifantakis.lib.ksafe.KSafe
import eu.anifantakis.lib.ksafe.KSafeConfig

/**
 * Platform entry point for constructing the [KSafe] instance used by [AuthTokenStore].
 *
 * - On Android, an application [android.content.Context] must be supplied via
 *   [setAndroidContext] (called from `Application.onCreate`).
 * - On iOS and JVM the no-arg constructor is used.
 */
expect object KSafeFactory {
    fun create(): KSafe
}

internal fun defaultConfig(): KSafeConfig = KSafeConfig(
    appNamespace = "com.glootea.goquiz"
)
