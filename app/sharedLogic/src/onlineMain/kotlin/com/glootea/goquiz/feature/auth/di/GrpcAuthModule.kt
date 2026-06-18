package com.glootea.goquiz.feature.auth.di

import com.glootea.goquiz.core.di.AppScope
import com.glootea.goquiz.core.grpc.newAuthServiceClient
import com.glootea.goquiz.feature.auth.data.AuthTokenStore
import com.glootea.goquiz.proto.auth.AuthClient
import dev.zacsweers.metro.ContributesTo
import dev.zacsweers.metro.Provides

@ContributesTo(AppScope::class)
interface GrpcAuthModule {

    companion object {
        @Provides
        fun provideAuthClient(tokenStore: AuthTokenStore): AuthClient = newAuthServiceClient(tokenStore)
    }
}
