package com.glootea.goquiz

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import com.glootea.goquiz.core.di.AppGraph
import com.glootea.goquiz.core.di.AppGraphFactory

class MainActivity : ComponentActivity() {

    private val graph: AppGraph by lazy { AppGraphFactory.create() }

    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge()
        super.onCreate(savedInstanceState)
        setContent {
            AppContent(graph)
        }
    }
}
