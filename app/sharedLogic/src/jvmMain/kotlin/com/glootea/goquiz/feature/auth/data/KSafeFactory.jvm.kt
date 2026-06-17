package com.glootea.goquiz.feature.auth.data

import eu.anifantakis.lib.ksafe.KSafe

actual object KSafeFactory {
    actual fun create(): KSafe = KSafe(config = defaultConfig())
}
