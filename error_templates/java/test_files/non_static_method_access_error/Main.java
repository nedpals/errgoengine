public class Main {
    // Non-static method
    public void printMessage() {
        System.out.println("Hello, World!");
    }

    public static void main(String[] args) {
        // Attempt to call the non-static method without creating an object
        printMessage(); // This will result in an error
    }
}
