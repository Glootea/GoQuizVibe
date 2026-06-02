package com.glootea.goquiz.feature.auth.api

import com.glootea.goquiz.core.di.AppScope
import com.glootea.goquiz.core.http.parseErrorOrThrow
import com.glootea.goquiz.feature.auth.api.dto.AuthResponseDto
import com.glootea.goquiz.feature.auth.api.dto.LoginRequestDto
import com.glootea.goquiz.feature.auth.api.dto.MeResponseDto
import com.glootea.goquiz.feature.auth.api.dto.RegisterRequestDto
import dev.zacsweers.metro.ContributesBinding
import dev.zacsweers.metro.Inject
import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.request.get
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.bodyAsText
import io.ktor.http.ContentType
import io.ktor.http.HttpStatusCode
import io.ktor.http.contentType
import kotlinx.serialization.json.Json

@ContributesBinding(AppScope::class)
@Inject
class AuthApiImpl(
    private val client: HttpClient,
    private val json: Json
) : AuthApi {

    override suspend fun login(request: LoginRequestDto): AuthResponseDto {
        val response = client.post("/api/v1/auth/login") {
            contentType(ContentType.Application.Json)
            setBody(json.encodeToString(LoginRequestDto.serializer(), request))
        }
        response.parseErrorOrThrow()
        return response.body()
    }

    override suspend fun register(request: RegisterRequestDto): AuthResponseDto {
        val response = client.post("/api/v1/auth/register") {
            contentType(ContentType.Application.Json)
            setBody(json.encodeToString(RegisterRequestDto.serializer(), request))
        }
        response.parseErrorOrThrow()
        return response.body()
    }

    override suspend fun logout() {
        val response = client.post("/api/v1/auth/logout")
        response.parseErrorOrThrow()
    }

    override suspend fun me(): MeResponseDto? {
        val response = client.get("/api/v1/auth/me")
        if (response.status == HttpStatusCode.Unauthorized) {
            return null
        }
        response.parseErrorOrThrow()
        val raw = response.bodyAsText()
        if (raw.isBlank()) return null
        return json.decodeFromString(MeResponseDto.serializer(), raw)
    }
}
