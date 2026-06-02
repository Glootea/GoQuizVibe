package com.glootea.goquiz.feature.auth

import com.glootea.goquiz.feature.auth.domain.usecase.LoginUseCase
import kotlinx.coroutines.test.runTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class LoginUseCaseTest {

    @Test
    fun login_callsRepositoryWithTrimmedEmail() = runTest {
        val repo = FakeAuthRepository(loginResult = Result.success(TestUsers.student))
        val useCase = LoginUseCase(repo)

        useCase("  test@example.com  ", "secret")

        assertEquals("test@example.com", repo.lastLoginEmail)
        assertEquals("secret", repo.lastLoginPassword)
    }

    @Test
    fun login_rejectsBlankEmail() = runTest {
        val repo = FakeAuthRepository()
        val useCase = LoginUseCase(repo)
        assertFailsWith<IllegalArgumentException> { useCase("", "pwd") }
        assertFailsWith<IllegalArgumentException> { useCase("   ", "pwd") }
    }

    @Test
    fun login_rejectsEmptyPassword() = runTest {
        val repo = FakeAuthRepository()
        val useCase = LoginUseCase(repo)
        assertFailsWith<IllegalArgumentException> { useCase("a@b.c", "") }
    }

    @Test
    fun login_propagatesRepositoryError() = runTest {
        val repo = FakeAuthRepository(loginResult = Result.failure(IllegalStateException("bad")))
        val useCase = LoginUseCase(repo)
        assertFailsWith<IllegalStateException> { useCase("a@b.c", "pwd") }
    }
}
