public class Main {
    public static void main(String[] args) {
        AnotherClass anotherClass = new AnotherClass();
        // Attempting to access a private variable from another class
        int value = anotherClass.privateVariable;
        System.out.println(value);
    }
}

class AnotherClass {
    private int privateVariable = 10;
}
