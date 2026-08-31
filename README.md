# ⚒️ KeyForge

KeyForge is a Linux keyboard remapping and automation engine.

The project is designed around a low-latency event pipeline where keyboard
events are processed entirely in memory.

## Architecture

```text
Physical Keyboard
       │
       ▼
     evdev
       │
       ▼
   KeyEvent
       │
       ▼
 Mapping Engine
       │
       ├───────────────┐
       │               │
    mapped          unmapped
       │               │
       └───────┬───────┘
               ▼
            uinput
               │
               ▼
      Virtual Keyboard