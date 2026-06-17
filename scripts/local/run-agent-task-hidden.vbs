Option Explicit

Dim shell
Dim command
Dim powershellPath
Dim index

Set shell = CreateObject("WScript.Shell")
powershellPath = shell.ExpandEnvironmentStrings("%SystemRoot%") & "\System32\WindowsPowerShell\v1.0\powershell.exe"
command = QuoteArg(powershellPath) & " -NoProfile -NonInteractive -ExecutionPolicy Bypass -WindowStyle Hidden"

For index = 0 To WScript.Arguments.Count - 1
    command = command & " " & QuoteArg(WScript.Arguments(index))
Next

shell.Run command, 0, True

Function QuoteArg(ByVal value)
    QuoteArg = Chr(34) & Replace(value, Chr(34), Chr(92) & Chr(34)) & Chr(34)
End Function
