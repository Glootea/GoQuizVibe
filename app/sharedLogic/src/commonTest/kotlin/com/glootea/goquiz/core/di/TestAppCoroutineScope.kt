package com.glootea.goquiz.core.di

import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.test.TestScope

class TestAppCoroutineScope(testScope: TestScope) {
    val value: CoroutineScope = testScope
}
