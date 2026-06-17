package com.glootea.goquiz.core.grpc

import com.glootea.goquiz.feature.auth.data.AuthTokenStore
import com.glootea.goquiz.proto.auth.AuthClient
import com.glootea.goquiz.proto.auth.GrpcAuthClient
import com.squareup.wire.GrpcClient
import kotlinx.coroutines.runBlocking
import okhttp3.Interceptor
import okhttp3.OkHttpClient
import okhttp3.Protocol
import okhttp3.Response
import java.util.concurrent.TimeUnit

object AuthConfig {
    const val HOST: String = "localhost"
    const val PORT: Int = 9100
    const val USE_TLS: Boolean = false
    val GRPC_URL: String = "${if (USE_TLS) "https" else "http"}://$HOST:$PORT"
}

fun newAuthServiceClient(tokenStore: AuthTokenStore): AuthClient {
    val tokenProvider = { runBlocking { tokenStore.accessToken() } }
    val okHttp = OkHttpClient.Builder()
        .protocols(listOf(Protocol.HTTP_2, Protocol.HTTP_1_1))
        .connectTimeout(15, TimeUnit.SECONDS)
        .readTimeout(30, TimeUnit.SECONDS)
        .addInterceptor(BearerTokenInterceptor(tokenProvider))
        .build()

    val grpcClient = GrpcClient.Builder()
        .client(okHttp)
        .baseUrl(AuthConfig.GRPC_URL)
        .build()

    return GrpcAuthClient(grpcClient)
}

private class BearerTokenInterceptor(
    private val tokenProvider: () -> String?
) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()
        val token = tokenProvider()
        val newRequest = if (!token.isNullOrBlank()) {
            request.newBuilder().addHeader("authorization", "Bearer $token").build()
        } else {
            request
        }
        return chain.proceed(newRequest)
    }
}
