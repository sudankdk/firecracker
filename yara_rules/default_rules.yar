rule Suspicious_EXE_Header {
    meta:
        description = "Detects Windows executable files"
        severity = "info"
    strings:
        $mz = "MZ"
    condition:
        $mz at 0
}

rule Potential_Ransomware_Keywords {
    meta:
        description = "Contains ransomware-related keywords"
        severity = "high"
    strings:
        $ransom1 = "encrypted" nocase
        $ransom2 = "bitcoin" nocase
        $ransom3 = "payment" nocase
        $ransom4 = "decrypt" nocase
    condition:
        3 of them
}

rule Suspicious_Shell_Commands {
    meta:
        description = "Shell command execution patterns"
        severity = "medium"
    strings:
        $cmd1 = "cmd.exe" nocase
        $cmd2 = "powershell" nocase
        $exec1 = "exec" nocase
        $exec2 = "system" nocase
    condition:
        any of ($cmd*) and any of ($exec*)
}

rule Crypto_Mining_Indicators {
    meta:
        description = "Cryptocurrency mining patterns"
        severity = "high"
    strings:
        $crypto1 = "monero" nocase
        $crypto2 = "mining" nocase
        $crypto3 = "stratum" nocase
    condition:
        2 of them
}

rule Keylogger_Indicators {
    meta:
        description = "Keylogger behavior patterns"
        severity = "critical"
    strings:
        $key1 = "GetAsyncKeyState" nocase
        $key2 = "keypress" nocase
        $log = "log" nocase
    condition:
        any of ($key*) and $log
}