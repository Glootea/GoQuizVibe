package com.glootea.goquiz.feature.auth.data

import com.glootea.goquiz.core.http.ApiException
import com.glootea.goquiz.feature.auth.api.AuthApi
import com.glootea.goquiz.feature.auth.api.dto.AuthResponseDto
import com.glootea.goquiz.feature.auth.api.dto.LoginRequestDto
import com.glootea.goquiz.feature.auth.api.dto.MeResponseDto
import com.glootea.goquiz.feature.auth.api.dto.RegisterRequestDto
import com.glootea.goquiz.feature.auth.api.dto.UserDto
import dev.zacsweers.metro.Inject
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock

@Inject
class LocalAuthApi : AuthApi {

    private val mutex = Mutex()
    private val users = mutableMapOf<String, StoredUser>(
        "teacher@example.com" to StoredUser("1", "Test teacher", "teacher@example.com", "teacher", "teacher123"),
        "student@example.com" to StoredUser("2", "Test student", "student@example.com", "student", "student123"),

    )
    private var currentUser: UserDto? = null

    override suspend fun login(request: LoginRequestDto): AuthResponseDto = mutex.withLock {
        val stored = users[request.email]
            ?: throw ApiException(code = "UNAUTHORIZED", message = "invalid credentials")
        if (stored.password != request.password) {
            throw ApiException(code = "UNAUTHORIZED", message = "invalid credentials")
        }
        currentUser = stored.toUserDto()
        AuthResponseDto(
            user = stored.toUserDto(),
            access_token = "local-${stored.id}",
            expires_at = "0"
        )
    }

    override suspend fun register(request: RegisterRequestDto): AuthResponseDto = mutex.withLock {
        if (request.name.isBlank() || request.email.isBlank() || request.password.isBlank()) {
            throw ApiException(code = "VALIDATION", message = "name, email and password must not be blank")
        }
        if (users.containsKey(request.email)) {
            throw ApiException(code = "CONFLICT", message = "user with this email already exists")
        }
        val stored = StoredUser(
            id = newId(),
            name = request.name,
            email = request.email,
            role = "student",
            password = request.password
        )
        users[stored.email] = stored
        currentUser = stored.toUserDto()
        AuthResponseDto(
            user = stored.toUserDto(),
            access_token = "local-${stored.id}",
            expires_at = "0"
        )
    }

    override suspend fun logout(): Unit = mutex.withLock {
        currentUser = null
    }

    override suspend fun me(): MeResponseDto? = mutex.withLock {
        currentUser?.let { MeResponseDto(user = it) }
    }

    private fun StoredUser.toUserDto(): UserDto =
        UserDto(id = id, name = name, email = email, role = role)

    private fun newId(): String =
        "local-${System.nanoTime()}-${(0 until Int.MAX_VALUE).random()}"

    private data class StoredUser(
        val id: String,
        val name: String,
        val email: String,
        val role: String,
        val password: String
    )
}
