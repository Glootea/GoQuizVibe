package com.glootea.goquiz.feature.auth

import com.glootea.goquiz.core.http.ApiException
import com.glootea.goquiz.feature.auth.api.dto.LoginRequestDto
import com.glootea.goquiz.feature.auth.api.dto.RegisterRequestDto
import com.glootea.goquiz.feature.auth.data.LocalAuthApi
import kotlinx.coroutines.test.runTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertNotNull
import kotlin.test.assertNull

class LocalAuthApiTest {

    @Test
    fun register_persistsUser_andLoginSucceedsAfter() = runTest {
        val api = LocalAuthApi()
        val registered = api.register(
            RegisterRequestDto(name = "Alice", email = "alice@local", password = "secret-1")
        )
        assertEquals("Alice", registered.user.name)
        assertEquals("alice@local", registered.user.email)
        assertEquals("student", registered.user.role)
        assertEquals("local-${registered.user.id}", registered.access_token)

        val loggedIn = api.login(LoginRequestDto(email = "alice@local", password = "secret-1"))
        assertEquals(registered.user.id, loggedIn.user.id)
    }

    @Test
    fun login_returnsUser_whenCredentialsValid() = runTest {
        val api = LocalAuthApi()
        api.register(RegisterRequestDto(name = "Bob", email = "bob@local", password = "secret-2"))

        val result = api.login(LoginRequestDto(email = "bob@local", password = "secret-2"))
        assertEquals("Bob", result.user.name)
        assertEquals("bob@local", result.user.email)
    }

    @Test
    fun login_throwsUnauthorized_whenEmailUnknown() = runTest {
        val api = LocalAuthApi()
        val ex = assertFailsWith<ApiException> {
            api.login(LoginRequestDto(email = "ghost@local", password = "any"))
        }
        assertEquals("UNAUTHORIZED", ex.code)
    }

    @Test
    fun login_throwsUnauthorized_whenPasswordMismatch() = runTest {
        val api = LocalAuthApi()
        api.register(RegisterRequestDto(name = "Cara", email = "cara@local", password = "right-pwd"))

        val ex = assertFailsWith<ApiException> {
            api.login(LoginRequestDto(email = "cara@local", password = "wrong-pwd"))
        }
        assertEquals("UNAUTHORIZED", ex.code)
    }

    @Test
    fun register_throwsConflict_whenEmailTaken() = runTest {
        val api = LocalAuthApi()
        api.register(RegisterRequestDto(name = "Dan", email = "dan@local", password = "pwd"))

        val ex = assertFailsWith<ApiException> {
            api.register(RegisterRequestDto(name = "Dan", email = "dan@local", password = "pwd"))
        }
        assertEquals("CONFLICT", ex.code)
    }

    @Test
    fun register_throwsValidation_whenFieldsBlank() = runTest {
        val api = LocalAuthApi()
        val ex = assertFailsWith<ApiException> {
            api.register(RegisterRequestDto(name = "", email = "e@local", password = "pwd"))
        }
        assertEquals("VALIDATION", ex.code)
    }

    @Test
    fun me_returnsCurrentUser_afterLogin() = runTest {
        val api = LocalAuthApi()
        api.register(RegisterRequestDto(name = "Eve", email = "eve@local", password = "pwd"))

        val me = api.me()
        assertNotNull(me)
        assertEquals("Eve", me.user.name)
    }

    @Test
    fun me_returnsNull_whenNeverLoggedIn() = runTest {
        val api = LocalAuthApi()
        assertNull(api.me())
    }

    @Test
    fun logout_clearsCurrentUser_butKeepsStoredCredentials() = runTest {
        val api = LocalAuthApi()
        api.register(RegisterRequestDto(name = "Finn", email = "finn@local", password = "pwd"))
        assertNotNull(api.me())

        api.logout()
        assertNull(api.me())

        val loggedIn = api.login(LoginRequestDto(email = "finn@local", password = "pwd"))
        assertEquals("Finn", loggedIn.user.name)
    }
}
