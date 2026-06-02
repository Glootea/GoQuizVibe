package com.glootea.goquiz.feature.auth

import com.glootea.goquiz.feature.auth.domain.usecase.LogoutUseCase
import com.glootea.goquiz.feature.auth.domain.usecase.ObserveMeUseCase
import com.glootea.goquiz.feature.auth.presentation.AuthState
import com.glootea.goquiz.feature.auth.presentation.AuthStateHolder
import com.glootea.goquiz.feature.auth.presentation.LoginViewModel
import com.glootea.goquiz.feature.auth.presentation.RegisterViewModel
import com.glootea.goquiz.feature.auth.domain.usecase.LoginUseCase
import com.glootea.goquiz.feature.auth.domain.usecase.RegisterUseCase
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

    private fun CoroutineScope.holder(repo: com.glootea.goquiz.feature.auth.domain.repository.AuthRepository) =
        AuthStateHolder(ObserveMeUseCase(repo), LogoutUseCase(repo), this)

    @Test
    fun loginViewModel_submit_succeedsAndUpdatesStateHolder() = runTest {
        val repo = FakeAuthRepository(loginResult = Result.success(TestUsers.student))
        val stateHolder = backgroundScope.holder(repo)
        val vm = LoginViewModel(LoginUseCase(repo), stateHolder)

        vm.onEmailChange("a@b.c")
        vm.onPasswordChange("secret")
        vm.submit()
        advanceUntilIdle()

        assertFalse(vm.state.value.isLoading)
        assertNull(vm.state.value.error)
        assertEquals(AuthState.Authenticated(TestUsers.student), stateHolder.state.value)
    }

    @Test
    fun loginViewModel_submit_setsErrorOnFailure() = runTest {
        val repo = FakeAuthRepository(loginResult = Result.failure(IllegalStateException("bad creds")))
        val stateHolder = backgroundScope.holder(repo)
        val vm = LoginViewModel(LoginUseCase(repo), stateHolder)

        vm.onEmailChange("a@b.c")
        vm.onPasswordChange("secret")
        vm.submit()
        advanceUntilIdle()

        assertEquals(false, vm.state.value.isLoading)
        assertEquals("bad creds", vm.state.value.error)
        assertEquals(AuthState.Unknown, stateHolder.state.value)
    }

    @Test
    fun loginViewModel_submit_doesNothingWhenDisabled() = runTest {
        val repo = FakeAuthRepository()
        val stateHolder = backgroundScope.holder(repo)
        val vm = LoginViewModel(LoginUseCase(repo), stateHolder)

        vm.submit()
        advanceUntilIdle()

        assertNull(repo.lastLoginEmail)
    }

    @Test
    fun registerViewModel_submit_succeedsForValidInput() = runTest {
        val repo = FakeAuthRepository(registerResult = Result.success(TestUsers.student))
        val stateHolder = backgroundScope.holder(repo)
        val vm = RegisterViewModel(RegisterUseCase(repo), stateHolder)

        vm.onNameChange("Bob")
        vm.onEmailChange("bob@x.com")
        vm.onPasswordChange("secret123")
        assertTrue(vm.state.value.isSubmitEnabled)
        vm.submit()
        advanceUntilIdle()

        assertEquals(AuthState.Authenticated(TestUsers.student), stateHolder.state.value)
        assertNull(vm.state.value.error)
    }

    @Test
    fun registerViewModel_submit_disabledForShortPassword() = runTest {
        val repo = FakeAuthRepository()
        val vm = RegisterViewModel(RegisterUseCase(repo), backgroundScope.holder(repo))

        vm.onNameChange("Bob")
        vm.onEmailChange("bob@x.com")
        vm.onPasswordChange("12345")

        assertFalse(vm.state.value.isSubmitEnabled)
        vm.submit()
        assertNull(repo.lastRegisterEmail)
    }
}
