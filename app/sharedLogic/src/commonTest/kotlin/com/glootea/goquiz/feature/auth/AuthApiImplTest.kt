package com.glootea.goquiz.feature.auth

import com.glootea.goquiz.core.http.ApiException
import com.glootea.goquiz.feature.auth.api.AuthApiImpl
import com.glootea.goquiz.feature.auth.api.dto.LoginRequestDto
import com.glootea.goquiz.feature.auth.api.dto.RegisterRequestDto
import io.ktor.client.HttpClient
import io.ktor.client.engine.mock.MockEngine
import io.ktor.client.engine.mock.respond
import io.ktor.client.engine.mock.respondError
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.http.HttpStatusCode
import io.ktor.http.headersOf
import io.ktor.serialization.kotlinx.json.json
import io.ktor.utils.io.ByteReadChannel
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.Json
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertNotNull

class AuthApiImplTest {

    private val json = Json { ignoreUnknownKeys = true; isLenient = true }

    private fun clientFor(handler: io.ktor.client.engine.mock.MockRequestHandler): HttpClient {
        val engine = MockEngine(handler)
        return HttpClient(engine) {
            install(ContentNegotiation) { json(json) }
        }
    }

    @Test
    fun login_returnsAuthResponseOnSuccess() = runTest {
        val body = """
            {"user":{"id":"1","name":"u","email":"a@b.c","role":"student"},
             "access_token":"token","expires_at":"0"}
        """.trimIndent()
        val client = clientFor { request ->
            assertEquals("/api/v1/auth/login", request.url.encodedPath)
            respond(
                content = ByteReadChannel(body),
                status = HttpStatusCode.OK,
                headers = headersOf("Content-Type", "application/json")
            )
        }
        val api = AuthApiImpl(client, json)

        val response = api.login(LoginRequestDto("a@b.c", "secret"))

        assertEquals("1", response.user.id)
        assertEquals("token", response.access_token)
    }

    @Test
    fun login_throwsApiException_onUnauthorized() = runTest {
        val body = """{"error":{"code":"INVALID_CREDENTIALS","message":"bad"}}""".trimIndent()
        val client = clientFor {
            respond(
                content = ByteReadChannel(body),
                status = HttpStatusCode.Unauthorized,
                headers = headersOf("Content-Type", "application/json")
            )
        }
        val api = AuthApiImpl(client, json)

        val exception = assertFailsWith<ApiException> { api.login(LoginRequestDto("a@b.c", "secret")) }
        assertEquals(HttpStatusCode.Unauthorized, exception.status)
        assertEquals("INVALID_CREDENTIALS", exception.code)
    }

    @Test
    fun register_throwsApiException_onConflict() = runTest {
        val body = """{"error":{"code":"CONFLICT","message":"email exists"}}""".trimIndent()
        val client = clientFor {
            respond(
                content = ByteReadChannel(body),
                status = HttpStatusCode.Conflict,
                headers = headersOf("Content-Type", "application/json")
            )
        }
        val api = AuthApiImpl(client, json)

        val exception = assertFailsWith<ApiException> { api.register(RegisterRequestDto("u", "a@b.c", "secret")) }
        assertEquals(HttpStatusCode.Conflict, exception.status)
    }

    @Test
    fun me_returnsNull_onUnauthorized() = runTest {
        val client = clientFor {
            respond(
                content = ByteReadChannel(""),
                status = HttpStatusCode.Unauthorized
            )
        }
        val api = AuthApiImpl(client, json)
        assertEquals(null, api.me())
    }

    @Test
    fun me_returnsDto_onSuccess() = runTest {
        val body = """{"user":{"id":"1","name":"u","email":"a@b.c","role":"student"}}""".trimIndent()
        val client = clientFor {
            respond(
                content = ByteReadChannel(body),
                status = HttpStatusCode.OK,
                headers = headersOf("Content-Type", "application/json")
            )
        }
        val api = AuthApiImpl(client, json)
        val me = api.me()
        assertNotNull(me)
        assertEquals("1", me.user.id)
    }
}
