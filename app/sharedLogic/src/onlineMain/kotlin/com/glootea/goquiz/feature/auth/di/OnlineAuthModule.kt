package com.glootea.goquiz.feature.auth.di

import com.glootea.goquiz.core.di.AppScope
import com.glootea.goquiz.feature.auth.api.AuthApi
import com.glootea.goquiz.feature.auth.api.AuthGrpcApiImpl
import dev.zacsweers.metro.ContributesTo
import dev.zacsweers.metro.Provides

@ContributesTo(AppScope::class)
interface OnlineAuthModule {

    companion object {
        @Provides
        fun provideAuthApi(grpc: AuthGrpcApiImpl): AuthApi = grpc
    }
}
