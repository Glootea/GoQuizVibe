package com.glootea.goquiz.feature.auth

import com.glootea.goquiz.feature.auth.api.AuthApi
import com.glootea.goquiz.feature.auth.api.dto.AuthResponseDto
import com.glootea.goquiz.feature.auth.api.dto.LoginRequestDto
import com.glootea.goquiz.feature.auth.api.dto.MeResponseDto
import com.glootea.goquiz.feature.auth.api.dto.RegisterRequestDto
import com.glootea.goquiz.feature.auth.api.dto.UserDto
import com.glootea.goquiz.feature.auth.data.AuthRepositoryImpl
import com.glootea.goquiz.feature.auth.domain.model.Role
import kotlinx.coroutines.test.runTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertNull

class FakeAuthApi(
    private val userByEmail: Map<String, UserDto> = emptyMap()
) : AuthApi {

    var lastLogin: LoginRequestDto? = null
    var lastRegister: RegisterRequestDto? = null
    var shouldFail: Boolean = false

    override suspend fun login(request: LoginRequestDto): AuthResponseDto {
        lastLogin = request
        if (shouldFail) error("login failed")
        val user = userByEmail[request.email]
            ?: error("user not found")
        return AuthResponseDto(
            user = user,
            access_token = "token-${request.email}",
            expires_at = "0"
        )
    }

    override suspend fun register(request: RegisterRequestDto): AuthResponseDto {
        lastRegister = request
        if (shouldFail) error("register failed")
        val user = UserDto(
            id = "id-${request.email}",
            name = request.name,
            email = request.email,
            role = "student"
        )
        return AuthResponseDto(
            user = user,
            access_token = "token-${request.email}",
            expires_at = "0"
        )
    }

    override suspend fun logout() = Unit

    override suspend fun me(): MeResponseDto? = null
}

class AuthRepositoryImplTest {

    @Test
    fun login_returnsDomainUserFromApi() = runTest {
        val api = FakeAuthApi(
            userByEmail = mapOf("a@b.c" to UserDto("1", "Alice", "a@b.c", "student"))
        )
        val repository = AuthRepositoryImpl(api)

        val user = repository.login("a@b.c", "secret")

        assertEquals("1", user.id)
        assertEquals("Alice", user.name)
        assertEquals("a@b.c", user.email)
        assertEquals(Role.Student, user.role)
        assertEquals("a@b.c", api.lastLogin?.email)
    }

    @Test
    fun register_passesThroughToApi() = runTest {
        val api = FakeAuthApi()
        val repository = AuthRepositoryImpl(api)

        val user = repository.register("Bob", "  bob@x.y  ", "pwd12345")

        assertEquals("Bob", user.name)
        assertEquals("  bob@x.y  ", user.email)
        assertEquals("  bob@x.y  ", api.lastRegister?.email)
        assertEquals("Bob", api.lastRegister?.name)
        assertEquals("pwd12345", api.lastRegister?.password)
    }

    @Test
    fun me_returnsNullWhenApiReturnsNull() = runTest {
        val api = FakeAuthApi()
        val repository = AuthRepositoryImpl(api)
        assertNull(repository.me())
    }

    @Test
    fun login_propagatesApiFailure() = runTest {
        val api = FakeAuthApi().also { it.shouldFail = true }
        val repository = AuthRepositoryImpl(api)
        assertFailsWith<IllegalStateException> { repository.login("a", "b") }
    }
}
