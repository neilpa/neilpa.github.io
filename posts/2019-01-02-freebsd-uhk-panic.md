# Fixing FreeBSD kernel panic with UHK keyboard

I've been using FreeBSD on and off as my desktop OS for a couple months (after ~5 years on OSX/macOS). Overall, I like the simplicity and design of the system. However, there's been a non-trivial amount of hacking and debugging to get some things working.

Most recently I discovered that my new [UHK keyboard][uhk] causes a kernel panic when booting. Early in the boot process I was greeted with the following message (stack addresses omitted).

```
ACPI APIC Table: <DELL   CBX3   >
panic: AP #1 (PHY# 1) failed!
cpuid = 0
KDB: stack backtrace:
#0 0x... at kdb_backtrace+0x67
#1 0x... at vpanic+0x177
#2 0x... at panic+0x43
#3 0x... at native_start_all_aps+0x344
#4 0x... at cpu_mp_start+0x2eb
#5 0x... at mp_start+0xa4
#6 0x... at mi_startup+0x118
#7 0x... at btext+0x2c
Uptime: 1s
```

Trying different USB ports didn't resolve the issue. Searching the forums and mailing lists turned up [this thread][0] with a similar stack trace and circumstance. Thankfully, the [penultimate][1] and [last message][2] had more details about the offending code, including an alternative fix:

> I think we should actually just remove the deassert INIT IPI entirely as I can find no reference in either the MP spec or otherwise that says that it should be used.  It is also ignored on all modern processors.

I'd yet to build a custom kernel, let alone patch it, but it was pretty simple thanks to the [handbook][3].

Checkout the sources matching your install if you don't already have a local copy

```
# svnlite checkout https://svn.FreeBSD.org/base/stable/11 /usr/src
```

After finding and patching the "broken" code in `sys/x86/x86/mp_x86.c`

```
# svnlite diff
Index: sys/x86/x86/mp_x86.c
===================================================================
--- sys/x86/x86/mp_x86.c	(revision 341355)
+++ sys/x86/x86/mp_x86.c	(working copy)
@@ -1045,13 +1045,6 @@
 	    APIC_LEVEL_ASSERT | APIC_DESTMODE_PHY | APIC_DELMODE_INIT, apic_id);
 	lapic_ipi_wait(100);
 
-	/* Explicitly deassert the INIT IPI. */
-	lapic_ipi_raw(APIC_DEST_DESTFLD | APIC_TRIGMOD_LEVEL |
-	    APIC_LEVEL_DEASSERT | APIC_DESTMODE_PHY | APIC_DELMODE_INIT,
-	    apic_id);
-
-	DELAY(10000);		/* wait ~10mS */
-
 	/*
 	 * next we do a STARTUP IPI: the previous INIT IPI might still be
 	 * latched, (P5 bug) this 1st STARTUP would then terminate
```

Now we can build, install and reboot into the newly fixed kernel.

```
# make -j 12 buildkernel KERNCONF=GENERIC
# make installworld KERNCONF=GENERIC
# shutdown -r now
```

And sure enough that fixed the issue. This was a quick and fun way to try out some kernel hacking.

[uhk]: https://ultimatehackingkeyboard.com/
[0]: https://lists.freebsd.org/pipermail/freebsd-stable/2012-June/068490.html
[1]: https://lists.freebsd.org/pipermail/freebsd-stable/2012-August/069125.html
[2]: https://lists.freebsd.org/pipermail/freebsd-stable/2012-August/069151.html
[3]: https://www.freebsd.org/doc/handbook/kernelconfig-building.html
