public class InvalidMethodExample {
    public static void main(String[] args) {
        int result = addNumbers(5, 10);
        System.out.println("Sum: " + result);
    }

    // Invalid method declaration with missing return type
    addNumbers(int a, int b) {
        return a + b;
    }
}