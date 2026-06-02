package com.glootea.goquiz.feature.auth.di

import com.glootea.goquiz.core.di.AppCoroutineScope
import com.glootea.goquiz.core.di.AppScope
import com.glootea.goquiz.core.http.ApiConfig
import com.glootea.goquiz.core.http.HttpClientFactory
import dev.zacsweers.metro.ContributesTo
import dev.zacsweers.metro.Provides
import io.ktor.client.HttpClient
import kotlinx.coroutines.CoroutineScope
import kotlinx.serialization.json.Json

@ContributesTo(AppScope::class)
interface AuthModule {

    companion object {
        @Provides
        fun provideHttpClient(): HttpClient = HttpClientFactory.create(ApiConfig)

        @Provides
        fun provideJson(): Json = Json {
            ignoreUnknownKeys = true
            isLenient = true
            explicitNulls = false
            encodeDefaults = true
        }

        @Provides
        fun provideCoroutineScope(appScope: AppCoroutineScope): CoroutineScope = appScope.value
    }
}
