package com.glootea.goquiz.feature.auth

import com.glootea.goquiz.feature.auth.domain.usecase.LogoutUseCase
import com.glootea.goquiz.feature.auth.domain.usecase.ObserveMeUseCase
import kotlinx.coroutines.test.runTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull

class LogoutAndObserveMeUseCaseTest {

    @Test
    fun logout_invokesRepository() = runTest {
        val repo = FakeAuthRepository()
        val useCase = LogoutUseCase(repo)
        useCase()
        assertEquals(1, repo.logoutCalled)
    }

    @Test
    fun observeMe_returnsUserFromRepository() = runTest {
        val repo = FakeAuthRepository().also { it.meResult = TestUsers.student }
        val useCase = ObserveMeUseCase(repo)
        assertEquals(TestUsers.student, useCase())
        assertEquals(1, repo.meCalled)
    }

    @Test
    fun observeMe_returnsNullWhenNoUser() = runTest {
        val repo = FakeAuthRepository()
        val useCase = ObserveMeUseCase(repo)
        assertNull(useCase())
    }
}
