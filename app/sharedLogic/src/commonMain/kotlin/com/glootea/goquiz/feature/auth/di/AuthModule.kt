package com.glootea.goquiz.feature.auth.di

import com.glootea.goquiz.core.di.AppCoroutineScope
import com.glootea.goquiz.core.di.AppScope
import com.glootea.goquiz.feature.auth.data.AuthTokenStore
import com.glootea.goquiz.feature.auth.data.KSafeFactory
import dev.zacsweers.metro.ContributesTo
import dev.zacsweers.metro.Provides
import eu.anifantakis.lib.ksafe.KSafe
import kotlinx.coroutines.CoroutineScope

@ContributesTo(AppScope::class)
interface AuthModule {

    companion object {
        @Provides
        fun provideKSafe(): KSafe = KSafeFactory.create()

        @Provides
        fun provideAuthTokenStore(ksafe: KSafe): AuthTokenStore = AuthTokenStore(ksafe)

        @Provides
        fun provideCoroutineScope(appScope: AppCoroutineScope): CoroutineScope = appScope.value
    }
}
