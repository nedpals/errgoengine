# java / NullPointerException

Your program try to access or manipulate an object reference that is currently pointing to `null`, meaning it doesn't refer to any actual object in memory. This typically happens when you forget to initialize an object before using it, or when you try to access an object that hasn't been properly assigned a value. 

## How to solve this issue?
- Check for any code that has `null` values
- Replace the value with a non-nullable value.

### Check for any code that has `null` values
Determine any variables that caused the errors. Based on the error we received, the error pointed to `ShouldBeNull.java` at line 7.

```java
    if (test.toUpperCase() == "123") {  
```

Your code accessed a method from a variable `test` which is at line 3.

```java
    String test = null;
```

### Replace the value with a non-nullable value
We can replace the value in line 3 with something like this:
```java
    String test = "";
```

Or add a check in line 7 to ensure that it is only executed if the value of the variable is not null.

```java
    if (test != null && test.toUpperCase() == "123") { 
```