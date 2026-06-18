package com.glootea.goquiz

import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.window.ComposeViewport
import com.glootea.goquiz.core.di.AppGraphFactory

@OptIn(ExperimentalComposeUiApi::class)
fun main() {
    val graph = AppGraphFactory.create()
    ComposeViewport {
        AppContent(graph)
    }
}
