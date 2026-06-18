package com.glootea.goquiz.feature.auth.api

import com.glootea.goquiz.core.di.AppScope
import com.glootea.goquiz.core.http.ApiException
import com.glootea.goquiz.feature.auth.api.dto.AuthResponseDto
import com.glootea.goquiz.feature.auth.api.dto.LoginRequestDto
import com.glootea.goquiz.feature.auth.api.dto.MeResponseDto
import com.glootea.goquiz.feature.auth.api.dto.RegisterRequestDto
import com.glootea.goquiz.feature.auth.api.dto.UserDto
import com.glootea.goquiz.feature.auth.data.AuthTokenStore
import com.glootea.goquiz.proto.auth.AuthClient
import com.glootea.goquiz.proto.auth.AuthResponse
import com.glootea.goquiz.proto.auth.LoginRequest
import com.glootea.goquiz.proto.auth.LogoutRequest
import com.glootea.goquiz.proto.auth.MeRequest
import com.glootea.goquiz.proto.auth.RegisterRequest
import com.glootea.goquiz.proto.auth.User
import com.squareup.wire.GrpcException
import com.squareup.wire.GrpcStatus
import dev.zacsweers.metro.Inject
import io.ktor.http.HttpStatusCode
import kotlinx.coroutines.runBlocking

@Inject
class AuthGrpcApiImpl(
    private val client: AuthClient,
    private val tokenStore: AuthTokenStore
) : AuthApi {

    override suspend fun login(request: LoginRequestDto): AuthResponseDto {
        val response = runCall {
            client.Login().execute(
                LoginRequest(
                    email = request.email,
                    password = request.password
                )
            )
        }
        saveSession(response)
        return response.toDto()
    }

    override suspend fun register(request: RegisterRequestDto): AuthResponseDto {
        val response = runCall {
            client.Register().execute(
                RegisterRequest(
                    name = request.name,
                    email = request.email,
                    password = request.password
                )
            )
        }
        saveSession(response)
        return response.toDto()
    }

    override suspend fun logout() {
        runCall { client.Logout().execute(LogoutRequest()) }
        tokenStore.clear()
    }

    override suspend fun me(): MeResponseDto? = try {
        val user = runCall { client.Me().execute(MeRequest()) }
        MeResponseDto(user = user.toDto())
    } catch (e: GrpcException) {
        if (e.grpcStatus == GrpcStatus.UNAUTHENTICATED) null
        else throw e.toApiException()
    }

    private suspend fun <T> runCall(block: suspend () -> T): T = try {
        block()
    } catch (e: GrpcException) {
        throw e.toApiException()
    } catch (e: java.io.IOException) {
        throw ApiException(
            code = "NETWORK",
            message = e.message ?: "gRPC call failed"
        )
    }

    private fun GrpcException.toApiException(): ApiException {
        val status = this.grpcStatus
        val httpStatus = when (status) {
            GrpcStatus.INVALID_ARGUMENT -> HttpStatusCode.BadRequest
            GrpcStatus.UNAUTHENTICATED -> HttpStatusCode.Unauthorized
            GrpcStatus.PERMISSION_DENIED -> HttpStatusCode.Forbidden
            GrpcStatus.NOT_FOUND -> HttpStatusCode.NotFound
            GrpcStatus.ALREADY_EXISTS -> HttpStatusCode.Conflict
            GrpcStatus.UNAVAILABLE -> HttpStatusCode.ServiceUnavailable
            GrpcStatus.DEADLINE_EXCEEDED -> HttpStatusCode.GatewayTimeout
            else -> HttpStatusCode.InternalServerError
        }
        val codeStr = when (status) {
            GrpcStatus.INVALID_ARGUMENT -> "BAD_REQUEST"
            GrpcStatus.UNAUTHENTICATED -> "UNAUTHORIZED"
            GrpcStatus.PERMISSION_DENIED -> "FORBIDDEN"
            GrpcStatus.NOT_FOUND -> "NOT_FOUND"
            GrpcStatus.ALREADY_EXISTS -> "CONFLICT"
            else -> "INTERNAL"
        }
        return ApiException(
            code = codeStr,
            message = this.grpcMessage ?: "gRPC call failed"
        )
    }

    private fun saveSession(response: AuthResponse) {
        val user = response.user ?: return
        if (user.id.isEmpty()) return
        runBlocking {
            tokenStore.put(
                token = response.access_token,
                userId = user.id,
                role = user.role
            )
        }
    }
}

private fun AuthResponse.toDto(): AuthResponseDto =
    AuthResponseDto(
        user = user?.let { UserDto(id = it.id, name = it.name, email = it.email, role = it.role) }
            ?: UserDto(id = "", name = "", email = "", role = ""),
        access_token = access_token,
        expires_at = expires_at.toString()
    )

private fun User.toDto(): UserDto =
    UserDto(id = id, name = name, email = email, role = role)
