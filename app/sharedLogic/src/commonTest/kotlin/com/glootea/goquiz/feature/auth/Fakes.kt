package com.glootea.goquiz.feature.auth

import com.glootea.goquiz.feature.auth.api.AuthApi
import com.glootea.goquiz.feature.auth.api.dto.AuthResponseDto
import com.glootea.goquiz.feature.auth.api.dto.LoginRequestDto
import com.glootea.goquiz.feature.auth.api.dto.MeResponseDto
import com.glootea.goquiz.feature.auth.api.dto.RegisterRequestDto
import com.glootea.goquiz.feature.auth.domain.model.Role
import com.glootea.goquiz.feature.auth.domain.model.User
import com.glootea.goquiz.feature.auth.domain.repository.AuthRepository

class FakeAuthRepository(
    private val loginResult: Result<User> = Result.failure(IllegalStateException("not set")),
    private val registerResult: Result<User> = Result.failure(IllegalStateException("not set"))
) : AuthRepository {

    var lastLoginEmail: String? = null
    var lastLoginPassword: String? = null
    var loginCalls: Int = 0
    var lastRegisterName: String? = null
    var lastRegisterEmail: String? = null
    var lastRegisterPassword: String? = null
    var registerCalls: Int = 0
    var logoutCalled: Int = 0
    var meResult: User? = null
    var meCalled: Int = 0

    override suspend fun login(email: String, password: String): User {
        lastLoginEmail = email
        lastLoginPassword = password
        loginCalls++
        return loginResult.getOrThrow()
    }

    override suspend fun register(name: String, email: String, password: String): User {
        lastRegisterName = name
        lastRegisterEmail = email
        lastRegisterPassword = password
        registerCalls++
        return registerResult.getOrThrow()
    }

    override suspend fun logout() {
        logoutCalled++
    }

    override suspend fun me(): User? {
        meCalled++
        return meResult
    }
}

object TestUsers {
    val student: User = User(
        id = "id-1",
        name = "Test User",
        email = "test@example.com",
        role = Role.Student
    )

    val teacher: User = User(
        id = "id-2",
        name = "Teacher",
        email = "teacher@example.com",
        role = Role.Teacher
    )
}

class NoopAuthApi : AuthApi {
    override suspend fun login(request: LoginRequestDto): AuthResponseDto =
        AuthResponseDto(
            user = com.glootea.goquiz.feature.auth.api.dto.UserDto(
                id = "1", name = "n", email = "e", role = "student"
            ),
            access_token = "t",
            expires_at = "0"
        )

    override suspend fun register(request: RegisterRequestDto): AuthResponseDto =
        AuthResponseDto(
            user = com.glootea.goquiz.feature.auth.api.dto.UserDto(
                id = "1", name = request.name, email = request.email, role = "student"
            ),
            access_token = "t",
            expires_at = "0"
        )

    override suspend fun logout() = Unit

    override suspend fun me(): MeResponseDto? = null
}
