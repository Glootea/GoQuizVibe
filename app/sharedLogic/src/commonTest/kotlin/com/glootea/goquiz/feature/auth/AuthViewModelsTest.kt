package com.glootea.goquiz.feature.auth

import com.glootea.goquiz.feature.auth.domain.repository.AuthRepository
import com.glootea.goquiz.feature.auth.presentation.AuthState
import com.glootea.goquiz.feature.auth.presentation.AuthStateHolder
import com.glootea.goquiz.feature.auth.presentation.LoginState
import com.glootea.goquiz.feature.auth.presentation.LoginViewModel
import com.glootea.goquiz.feature.auth.presentation.RegisterState
import com.glootea.goquiz.feature.auth.presentation.RegisterViewModel
import com.glootea.goquiz.feature.auth.presentation.isSubmitEnabled
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import kotlin.test.AfterTest
import kotlin.test.BeforeTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertIs
import kotlin.test.assertNull
import kotlin.test.assertTrue

@OptIn(kotlinx.coroutines.ExperimentalCoroutinesApi::class)
class AuthViewModelsTest {

    @BeforeTest
    fun setup() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    @AfterTest
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun CoroutineScope.holder(repo: AuthRepository) =
        AuthStateHolder(repo, this)

    @Test
    fun loginViewModel_submit_succeedsAndUpdatesStateHolder() = runTest {
        val repo = FakeAuthRepository(loginResult = Result.success(TestUsers.student))
        val stateHolder = backgroundScope.holder(repo)
        val vm = LoginViewModel(repo, stateHolder)

        vm.onEmailChange("a@b.c")
        vm.onPasswordChange("secret")
        vm.submit()
        advanceUntilIdle()

        assertIs<LoginState.Editing>(vm.state.value)
        assertEquals(AuthState.Authenticated(TestUsers.student), stateHolder.state.value)
    }

    @Test
    fun loginViewModel_submit_setsErrorOnFailure() = runTest {
        val repo = FakeAuthRepository(loginResult = Result.failure(IllegalStateException("bad creds")))
        val stateHolder = backgroundScope.holder(repo)
        val vm = LoginViewModel(repo, stateHolder)

        vm.onEmailChange("a@b.c")
        vm.onPasswordChange("secret")
        vm.submit()
        advanceUntilIdle()

        val state = vm.state.value
        assertIs<LoginState.Error>(state)
        assertEquals("bad creds", state.message)
        assertEquals(AuthState.Unknown, stateHolder.state.value)
    }

    @Test
    fun loginViewModel_submit_doesNothingWhenDisabled() = runTest {
        val repo = FakeAuthRepository()
        val stateHolder = backgroundScope.holder(repo)
        val vm = LoginViewModel(repo, stateHolder)

        vm.submit()
        advanceUntilIdle()

        assertNull(repo.lastLoginEmail)
    }

    @Test
    fun loginViewModel_submit_invokesRepositoryOnceWhenCalledTwiceInARow() = runTest {
        val repo = FakeAuthRepository(loginResult = Result.success(TestUsers.student))
        val stateHolder = backgroundScope.holder(repo)
        val vm = LoginViewModel(repo, stateHolder)

        vm.onEmailChange("a@b.c")
        vm.onPasswordChange("secret")
        vm.submit()
        vm.submit()

        advanceUntilIdle()

        assertEquals(1, repo.loginCalls)
    }

    @Test
    fun loginViewModel_trimsEmailOnSubmit() = runTest {
        val repo = FakeAuthRepository(loginResult = Result.success(TestUsers.student))
        val stateHolder = backgroundScope.holder(repo)
        val vm = LoginViewModel(repo, stateHolder)

        vm.onEmailChange("  test@example.com  ")
        vm.onPasswordChange("secret")
        vm.submit()
        advanceUntilIdle()

        assertEquals("test@example.com", repo.lastLoginEmail)
    }

    @Test
    fun registerViewModel_submit_succeedsForValidInput() = runTest {
        val repo = FakeAuthRepository(registerResult = Result.success(TestUsers.student))
        val stateHolder = backgroundScope.holder(repo)
        val vm = RegisterViewModel(repo, stateHolder)

        vm.onNameChange("Bob")
        vm.onEmailChange("bob@x.com")
        vm.onPasswordChange("secret123")
        assertTrue(vm.state.value.isSubmitEnabled)
        vm.submit()
        advanceUntilIdle()

        assertEquals(AuthState.Authenticated(TestUsers.student), stateHolder.state.value)
        assertIs<RegisterState.Editing>(vm.state.value)
    }

    @Test
    fun registerViewModel_submit_disabledForShortPassword() = runTest {
        val repo = FakeAuthRepository()
        val vm = RegisterViewModel(repo, backgroundScope.holder(repo))

        vm.onNameChange("Bob")
        vm.onEmailChange("bob@x.com")
        vm.onPasswordChange("12345")

        assertFalse(vm.state.value.isSubmitEnabled)
        vm.submit()
        assertNull(repo.lastRegisterEmail)
    }
}
