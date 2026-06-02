package com.glootea.goquiz.feature.auth

import com.glootea.goquiz.feature.auth.domain.usecase.LogoutUseCase
import com.glootea.goquiz.feature.auth.domain.usecase.ObserveMeUseCase
import com.glootea.goquiz.feature.auth.presentation.AuthState
import com.glootea.goquiz.feature.auth.presentation.AuthStateHolder
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

@OptIn(kotlinx.coroutines.ExperimentalCoroutinesApi::class)
class AuthStateHolderTest {

    private fun CoroutineScope.holder(repo: com.glootea.goquiz.feature.auth.domain.repository.AuthRepository) =
        AuthStateHolder(ObserveMeUseCase(repo), LogoutUseCase(repo), this)

    @Test
    fun initial_bootstrapAuthenticated_whenMeReturnsUser() = runTest {
        val repo = FakeAuthRepository().also { it.meResult = TestUsers.student }
        val stateHolder = holder(repo)

        advanceUntilIdle()

        val state = stateHolder.state.value
        assertTrue(state is AuthState.Authenticated)
        assertEquals(TestUsers.student, (state as AuthState.Authenticated).user)
    }

    @Test
    fun initial_bootstrapUnauthenticated_whenMeReturnsNull() = runTest {
        val repo = FakeAuthRepository()
        val stateHolder = holder(repo)

        advanceUntilIdle()

        assertEquals(AuthState.Unauthenticated, stateHolder.state.value)
    }

    @Test
    fun initial_bootstrapUnauthenticated_whenMeThrows() = runTest {
        val brokenRepo = object : com.glootea.goquiz.feature.auth.domain.repository.AuthRepository {
            override suspend fun login(email: String, password: String) = error("nope")
            override suspend fun register(name: String, email: String, password: String) = error("nope")
            override suspend fun logout() = Unit
            override suspend fun me(): com.glootea.goquiz.feature.auth.domain.model.User? {
                throw IllegalStateException("boom")
            }
        }
        val stateHolder = holder(brokenRepo)

        advanceUntilIdle()

        assertEquals(AuthState.Unauthenticated, stateHolder.state.value)
    }

    @Test
    fun setAuthenticated_updatesState() = runTest {
        val stateHolder = holder(FakeAuthRepository())
        stateHolder.setAuthenticated(TestUsers.student)
        assertEquals(AuthState.Authenticated(TestUsers.student), stateHolder.state.value)
    }

    @Test
    fun performLogout_clearsStateAndCallsRepository() = runTest {
        val repo = FakeAuthRepository()
        val stateHolder = holder(repo)
        stateHolder.setAuthenticated(TestUsers.student)

        stateHolder.performLogout()
        advanceUntilIdle()

        assertEquals(AuthState.Unauthenticated, stateHolder.state.value)
        assertEquals(1, repo.logoutCalled)
    }
}
