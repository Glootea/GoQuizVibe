package com.glootea.goquiz.feature.auth

import com.glootea.goquiz.feature.auth.domain.usecase.RegisterUseCase
import kotlinx.coroutines.test.runTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class RegisterUseCaseTest {

    @Test
    fun register_succeedsForValidInput() = runTest {
        val repo = FakeAuthRepository(registerResult = Result.success(TestUsers.student))
        val useCase = RegisterUseCase(repo)

        val user = useCase("Bob", "bob@x.com", "secret123")

        assertEquals(TestUsers.student, user)
        assertEquals("Bob", repo.lastRegisterName)
        assertEquals("bob@x.com", repo.lastRegisterEmail)
        assertEquals("secret123", repo.lastRegisterPassword)
    }

    @Test
    fun register_rejectsBlankName() = runTest {
        val repo = FakeAuthRepository()
        val useCase = RegisterUseCase(repo)
        assertFailsWith<IllegalArgumentException> { useCase("", "a@b.c", "secret1") }
    }

    @Test
    fun register_rejectsBlankEmail() = runTest {
        val repo = FakeAuthRepository()
        val useCase = RegisterUseCase(repo)
        assertFailsWith<IllegalArgumentException> { useCase("Name", "", "secret1") }
    }

    @Test
    fun register_rejectsShortPassword() = runTest {
        val repo = FakeAuthRepository()
        val useCase = RegisterUseCase(repo)
        assertFailsWith<IllegalArgumentException> { useCase("Name", "a@b.c", "12345") }
    }

    @Test
    fun register_propagatesRepositoryError() = runTest {
        val repo = FakeAuthRepository(registerResult = Result.failure(IllegalStateException("exists")))
        val useCase = RegisterUseCase(repo)
        assertFailsWith<IllegalStateException> { useCase("Bob", "bob@x.com", "secret1") }
    }
}
